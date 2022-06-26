import sys
import os
import requests
import base64
import json

# read file as bytestream
def read_file_b64(filename):
    with open(filename, 'rb') as f:
        data = f.read()
        return base64.b64encode(data)

def write_file_b64(filename, data):
    with open(filename, 'wb') as f:
        f.write(base64.b64decode(data))

req = {
    'x': 200,
    'y': 500,
    'data': read_file_b64('olaf.jpg'),
    'format': 'webp',
}

resp = requests.post('http://localhost:8080/limit', json=req)

img_json = json.loads(resp.text)

print("x: " + str(img_json["x"]))
print("y: " + str(img_json["y"]))
print("format: " + img_json["format"])

write_file_b64('olaf_out.webp', img_json["data"])

