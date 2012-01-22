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

package kanzi.entropy;

import kanzi.OutputBitStream;


// Null entropy encoder and decoder
// Pass through that writes the data directly to the bitstream
// Helpful to debug
public final class NullEntropyEncoder extends AbstractEncoder
{
    private final OutputBitStream bitstream;


    public NullEntropyEncoder(OutputBitStream bitstream)
    {
       if (bitstream == null)
          throw new NullPointerException("Invalid null bitstream parameter");

       this.bitstream = bitstream;
    }


    @Override
    public boolean encodeByte(byte val)
    {
        return (this.bitstream.writeBits(val, 8) == 8);
    }


    @Override
    public void dispose()
    {
    }


    @Override
    public OutputBitStream getBitStream()
    {
       return this.bitstream;
    }
}
