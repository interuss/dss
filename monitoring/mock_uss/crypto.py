from Crypto.PublicKey import RSA
from Crypto.Signature.pkcs1_15 import PKCS115_SigScheme
from Crypto.Hash import SHA256
import requests
import sys
import base64
import hashlib

covered_components = ["@method", "@path", "@query", "authorization", "content-type", "content-digest", "x-utm-jws-header"]
headers = {
    "authorization":"Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJob3N0LmRvY2tlci5pbnRlcm5hbCIsImV4cGlyZXNfaW4iOjE2NjY2NDczOTcsImlzcyI6ImR1bW15LmF1dGgiLCJqdGkiOiIzOTAyMDBjYS04M2M5LTQ0NjYtYmFhMi04ZDU3ZTE0YmMxNTMiLCJuYmYiOjE2NjY2NDM3OTcsInNjb3BlIjoidXRtLnN0cmF0ZWdpY19jb29yZGluYXRpb24iLCJzdWIiOiJucHN1MS50ZXN0aW5nLm5hc2EuZ292IiwidG9rZW5fdHlwZSI6ImJlYXJlciJ9.D8UPw4XzkLBXMmGiW7ZJVD8amQtgbtf78hHxQ5K41PJkXnPOwolLcMYyC4diQG4glEshblg3zVLmHQ1oZGUjM651eUa5YvkqVwaGWHFMpzeFMlyaRIwVEfT-l5cY9XfQRgdesqt6vviuMiH0-IEajhzA798oZYIEHgX_9FOqLwk", 
    "content-digest":"sha-512=:EPlfPvi5GIXLnzFiCkQHsirnCv1R3LzxYCawEDaLeaDltuYGt8h1U6cPCG2lmzUJj58rVWUc3BIWcTFatSqg6A==:", 
    "@method": "POST", 
    "x-utm-jws-header":'alg="RS256", typ="JOSE", kid="dEQrbd7DCiGBSa0B8Txt5k0xLoKgHbUi2bgqERe5ag8", x5u="http://host.docker.internal:10206/interop/.well-known/uas-traffic-management/UFAA_UssSignKey_pub_dev.der"',
    "content-type": "application/json", 
    "@query": "?", 
    "@path": "/mock/scd/uss/v1/operational_intents"
}

rsb = ""
for key in covered_components:
    rsb += "\"{}\": {}\n".format(
        key, headers[key]
    )
sig_params = "\"{}\": ({});created=1666654644774".format("@signature-params", " ".join(list(map(lambda comp: "\"{}\"".format(comp), covered_components))))
rsb = rsb + sig_params 
print("HASH OF RSB: " + base64.b64encode(hashlib.sha512(rsb.encode('utf-8')).digest()).decode('utf-8'))
print(repr(rsb))
from Crypto.Signature import pkcs1_15

rsb = bytes(rsb, 'utf-8')
signature = base64.b64decode("X/SQ6AHUWWB0sUDE29xrWHam9I2b7x4MzAU4L1N8oRi+fCFCwVAciowMzOdYTkAGB2rAnsFRNQGQemLctf/9cX8KJJ13GPB4Z+xP//w+0/ZqTJn815mo+EWObsRe+A94D7vYAdF2NXjjgcL3kTbSVHk9Vt36cWRH4xtpH39D7DZ7ckAJVvJeP5t4GDuzyy82580Tq8CCR0VKV3HOWHeFCpGKeXzXl6cONTJUr3rjwIh1la7fxyrIXVR18gwJW1nXl8vnYoDD+7OpkoBMGYPwj1x5PI5/T2syDVTf1OtQ9Swec0HjkFDNf318/tdCYJSQzzSpzgYsCIW07GVD3l5tqA==")
hash = SHA256.new(rsb)
public_key_content = requests.get("http://localhost:10206/interop/.well-known/uas-traffic-management/UFAA_UssSignKey_pub_dev.der").content
public_key = RSA.import_key(public_key_content)
with open("UFAA_UssSignKey_priv_dev.pem", "rb") as pkeyfile:
    private_key = RSA.import_key(pkeyfile.read())
    sig_2 = pkcs1_15.new(private_key).sign(hash)
    print(base64.b64encode(sig_2))
try:
    pkcs1_15.new(public_key).verify(hash, signature)
except ValueError:
    print("Not verified :(")
    sys.exit()
print("Verified!")