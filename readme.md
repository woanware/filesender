
# Why

I originally saw the implementation for [magic-wormhole](https://github.com/warner/magic-wormhole) and I liked the ability to send files from one computer to another. But...I didn't like having to rely on a third party for the relay aspect.

More research identified [croc](https://github.com/schollz/croc), which is ideally written in golang :-). Croc has a local mode, which in principle solved my requirements. But...I found it rather buggy, and whilst the author attempted to fix the bug quickly, I still found that it failed the next time I wanted to use it.

Yes I could have debugged and potentially solved the bug, but the application had numerous interconnected libraries, and the project uses golang modules which I haven't researched yet. Also the code had lots of functionality which I didn't require and wanted to avoid e.g. third party relay.

I decided to use Google Drive as the relay, as the auth is proven, it is commonly used, and there is an API for it.

Research on command line Google Drive applications found [gdrive](https://github.com/prasmussen/gdrive). But...it appeared unsupported, far too much functionality I didn't required, and numerous outstanding issues raised.

[skicka](https://github.com/google/skicka) was the next application found, which was more recently supported, had in-built crypto. But...again too much functionality and hadn't been upgraded to the newer V3 Google Drive API.

A newly created library for golang interaction with Google Drive called [gdriver](https://github.com/Eun/gdriver) was identified, and is used for the core Google Drive functionality, modified it slightly to permit custom meta data storage.

I used the crypto concept from [skicka](https://github.com/google/skicka), and the [mnemonicode](https://github.com/schollz/mnemonicode) concept from [magic-wormhole](https://github.com/warner/magic-wormhole)/[croc](https://github.com/schollz/croc).

# Usage

## Generate

The **generate** function/verb is used to create the crypto data used for the file storage. See the section (Encryption) below for more details.

## Unencrypted
```
./filesender send cat.jpg

filesender v0.0.1

Sending 63.9 KiB file: cat.jpg
 63.90 KiB / 63.90 KiB [==============================] 100.00% 3s

Code is: lola-first-fiber
On the other computer run: filesender r lola-first-fiber

```

```
./filesender receive lola-first-fiber

filesender v0.0.1

 63.90 KiB / 63.90 KiB [===============================] 100.00% 0s
Received cat.jpg file: 63.9 KiB

```

## Encrypted

Note the trailing **-e** parameter.

```
./filesender send cat.jpg -e

filesender v0.0.1

Enter password:

Sending 63.9 KiB file: cat.jpg
 63.92 KiB / 63.92 KiB [===============================] 100.00% 1s

Code is: copper-crater-arizona
On the other computer run: filesender r copper-crater-arizona
```

```
./filesender receive copper-crater-arizona

filesender v0.0.1

Enter password:

 63.90 KiB / 63.92 KiB [===============================]  99.98% 0s
Received cat.jpg file: 63.9 KiB

```

## Leave

Files are normally deleted after a successful download, but say you wanted to download the same file to multiple hosts, then you can specify the **-l** parameter, and the file will be left on Google Drive.

```
./filesender send cat.jpg -l
./filesender send cat.jpg -l -e
```

## Purge

The **purge** or **p** allows the user to remove (or purge) all existing filesender files from Google Drive. It checks the files meta data to ensure non filesender files are not removed e.g. if they mistakenly get put in the filesender folder

# Encryption

**Note**: This crypto concept is taken verbosely from [skicka](https://github.com/google/skicka)

1. When the **generate** function is run, **filesender** generates a random 32-byte
When the **generate** function is run, **filesender** generates a random 32-byte
config file.

2. The user's password is run through the PBKDF2 key derivation function, using 65536 iterations and the SHA-256 hash function to derive a 64-byte hash.

3. The first 32 bytes of the hash are hex encoded and is stored in **password_hash** field of the config file. These bytes are later used only to validate that the user has provided the correct password on subsequent runs of **filesender**.

4. A random 32-byte encryption key is generated (again with rand.Reader). This is the key that will actually be used for encryption and decryption of file contents.

5. A random 16-byte initialization vector is generated with rand.Reader. It is hex encoded and stored in the **encrypted_key_iv** field of the configuration file.

6. The encryption key from #4 is encrypted using the initialization vector from #5, using the second 32 bytes of the hash computed in #2 as the encryption key. The result is stored in the **encrypted_key** field of the config file.

Upon subsequent runs of filesender, the salt is loaded from the config file so that PBKDF2 can be used as in #2 above to hash the user's password. If the first 32 bytes of the passphrase hash match the stored bytes, then the second 32 bytes of the hash and the stored IV are used to decrypt the encrypted encryption key.

Given the encryption key, when a file is to be encrypted before being uploaded to Google Drive, filesender uses the key along with a fresh 16-byte initialization vector for each file to encrypt the file using AES-256.

The initialization vector is prepended to the file contents before upload. (Thus, encrypted files are 16 bytes larger on Google Drive than they are locally.)

The initialization vector is also stored hex-encoded as a Google Drive file Property with the name "IV". We store the initialization vector redundantly so that if one downloads the encrypted file contents, it's possible to decrypt the file using the file contents (and the key!) alone. Conversely, also having the IV available in a property makes it possible to encrypt the contents of a local version of a file without needing to download any of the contents of the corresponding file from Google Drive.

# Inspiration

- https://github.com/Eun/gdriver
- https://github.com/prasmussen/gdrive
- https://github.com/google/skicka
- https://github.com/schollz/croc

# TODO
- [ ] Configurable folder
- [x] Use Crypto config
- [x] Refactor e.g. small routines called by main functions, example get IV data from google file object, and do validation, and conversion #
- [x] Generate crypto config
- [x] Prompt for overwriting
- [x] Receive file
- [x] Delete file on receive
- [x] Have flag to leave file on google drive
- [x] Delete all files
