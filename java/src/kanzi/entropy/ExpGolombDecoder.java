/*
Copyright 2011-2013 Frederic Langlet
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

package kanzi.entropy;

import kanzi.InputBitStream;


// Exponential Golomb Coder
public final class ExpGolombDecoder extends AbstractDecoder
{
    private final boolean signed;
    private final InputBitStream bitstream;


    public ExpGolombDecoder(InputBitStream bitstream, boolean signed)
    {
        if (bitstream == null)
           throw new NullPointerException("Invalid null bitstream parameter");

        this.signed = signed;
        this.bitstream = bitstream;
    }


    public boolean isSigned()
    {
        return this.signed;
    }


    @Override
    public byte decodeByte()
    {
       int log2 = 0;

       // Decode unsigned
       while (this.bitstream.readBit() == 0)
           log2++;

       if (log2 == 0)
          return 0;
       
       long res = (1 << log2) - 1 + this.bitstream.readBits(log2);

       // Read signed if necessary
       if (this.signed == true)
       {
           // If res != 0, get the sign (1 for negative values)
           if (this.bitstream.readBit() == 1)
               return (byte) -res;
       }

       return (byte) res;
    }


    @Override
    public InputBitStream getBitStream()
    {
       return this.bitstream;
    }
}
