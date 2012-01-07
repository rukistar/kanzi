/*
Copyright 2011 Frederic Langlet
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
you may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
 */

package kanzi.bitstream;

import kanzi.BitStreamException;
import java.io.IOException;
import java.io.OutputStream;


/*package*/ final class OutputBitStream
{
   private final OutputStream os;
   private final byte[] buffer;
   private boolean closed;
   private int position;
   private int bitIndex;
   private long written;


   OutputBitStream(OutputStream os, int bufferSize)
   {
      this.os = os;
      this.buffer = new byte[bufferSize];
      this.bitIndex = 7;
   }


   // Processes the least significant bit of the input integer
   public synchronized boolean writeBit(int bit)
   {
      try
      {
         this.buffer[this.position] |= ((bit & 1) << this.bitIndex);
         this.bitIndex = (this.bitIndex + 7) & 7;
         this.written++;

         if (this.bitIndex == 7)
         {
            if (++this.position >= this.buffer.length)
               this.flush();
         }
      }
      catch (ArrayIndexOutOfBoundsException e)
      {
         throw new BitStreamException("Stream closed", BitStreamException.STREAM_CLOSED);
      }

      return true;
   }


   // 'length' must be max 64
   public synchronized int writeBits(long value, int length)
   {
      try
      {
         int remaining = length;

         // Pad the current position in buffer
         if (this.bitIndex != 7)
         {
            int idx = this.bitIndex;
            final int len = (remaining <= idx + 1) ? remaining : idx + 1;
            remaining -= len;
            final int bits = (int) ((value >> remaining) & ((1 << len) - 1));
            this.buffer[this.position] |= (bits << (idx + 1 - len));
            idx = (idx + 8 - len) & 7;
            this.written += len;
            this.bitIndex = idx;

            if (idx == 7)
            {
               if (++this.position >= this.buffer.length)
                  this.flush();
            }
         }

         if (this.bitIndex == 7)
         {
            // Progress byte by byte
            while (remaining >= 8)
            {
               remaining -= 8;
               this.buffer[this.position] = (byte) ((value >> remaining) & 0xFF);
               this.written += 8;

               if (++this.position >= this.buffer.length)
                  this.flush();
            }

            // Process remaining bits
            if (remaining > 0)
            {
               final int bits = (int) (value & ((1 << remaining) - 1));
               this.buffer[this.position] |= (bits << (8 - remaining));
               this.written += remaining;
               this.bitIndex -= remaining;
            }
         }

         return length;
      }
      catch (ArrayIndexOutOfBoundsException e)
      {
         throw new BitStreamException("Stream closed", BitStreamException.STREAM_CLOSED);
      }
   }


   public synchronized void flush() throws BitStreamException
   {
      if (this.isClosed() == true)
         throw new BitStreamException("Stream closed", BitStreamException.STREAM_CLOSED);

      try
      {
         if (this.position > 0)
         {
            final int len = this.position;
            this.os.write(this.buffer, 0, len);
            this.buffer[0] = (this.bitIndex != 7) ? this.buffer[this.position] : 0;

            for (int i=1; i<len; i++) // do not reset buffer[0]
               this.buffer[i] = 0;

            this.position = 0;
         }

         this.os.flush();
      }
      catch (IOException e)
      {
         throw new BitStreamException(e.getClass().getName(), BitStreamException.INPUT_OUTPUT);
      }
   }


   public synchronized void close() throws IOException
   {
      if (this.isClosed() == true)
         return;

      if ((this.written > 0) && (this.bitIndex != 7))
      {
         if (++this.position >= this.buffer.length)
            this.flush();

         this.written -= (7 - this.bitIndex);
         this.written += 8;
         this.bitIndex = 7;
      }

      // Adjust stream size to multiple of 8 bytes (to allow decoding of long)
      while ((this.written & 63) != 0)
      {
         // Pad with 0xFF
         this.buffer[this.position] = (byte) 0xFF;
         this.written += 8;

         if (++this.position >= this.buffer.length)
            this.flush();
      }

      this.flush();
      this.os.close();
      this.closed = true;

      // Force an exception on any write attempt
      this.position = this.buffer.length;
      this.bitIndex = 7;
   }


   public synchronized long written()
   {
      return this.written;
   }


   protected synchronized boolean isClosed()
   {
      return this.closed;
   }
}