#!/bin/sh -e

HASHES=$(curl -s https://etherscan.io/txs | pup 'span.address-tag a attr{href}' | grep -v address)

for h in $HASHES; do
    echo "Fetch tx $h"

    HTML=$(curl -s https://etherscan.io/getRawTx?tx=${h:4})
    TX=$(echo $HTML | pup '#ContentPlaceHolder1_divparitytrace .wordwrap text{}' | awk 'NR == 3')

    echo ${TX:2} > ${h:6}.hex
done

python - <<END
import codecs
import glob
import json
import os
import rlp
from rlp.sedes import *

class EthTx(rlp.Serializable):
    def __iter__(self):
        return iter([
            ('nonce', self.nonce),
                ('gasprice', str(self.gasprice)),
            ('startgas', self.startgas),
            ('to', "0x" + str(codecs.encode(self.to, "hex"),"utf-8")),
            ('value', str(self.value)),
            ('data', str(codecs.encode(self.data, "hex"), "utf-8")),
            ('v', self.v),
            ('r', str(self.r)),
            ('s', str(self.s))
            ])

    fields = [
            ('nonce', big_endian_int),
            ('gasprice', big_endian_int),
            ('startgas', big_endian_int),
            ('to', binary),
            ('value', big_endian_int),
            ('data', binary),
            ('v', big_endian_int),
            ('r', big_endian_int),
            ('s', big_endian_int)
            ]

if __name__ == "__main__":
    for g in glob.glob('./*.hex'):
        with open(g) as f:
            raw = f.read().rstrip()
            tx = codecs.decode(raw, "hex")
            js = rlp.decode(tx, EthTx)
            djs = dict(js)
            djs['txhash'] = g[2:-4]
            djs['hex'] = raw

            if js.v > 36:
                with open(g.replace("hex", "json"), "w") as jsout:
                    jsout.write(json.dumps(djs, sort_keys=True, indent=4, separators=(',', ': ')))

            os.remove(g)
END
