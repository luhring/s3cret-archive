# s3cret

Copy files to/from AWS S3 with automatic client-side encryption/decryption

Object metadata:

16 bytes      key id (UUID)
8 bytes       encrypted chunk data length (ECL) == chunkLength + Overhead (16 bytes)

Encrypted file format:

[] encrypted chunks

  8 bytes               index
  24 bytes              nonce
  *ECL bytes            data
