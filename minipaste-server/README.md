# minipaste

HTTP POST or PUT files with

```
$ curl -F'file=@foo.txt' https://example.com
$ curl https://example.com --upload-file foo.txt
```

HTTP DELETE files with
```
$ curl -XDELETE https://example.com/paste
```

Or use the client
```
$ nix run github:jktr/minipaste -- --server https://example.com foo.txt
$ cat foo.txt | nix run github:jktr/minipaste -- --server https://example.com
$ nix run github:jktr/minipaste -- --server https://example.com --delete
```

File URLs are valid for at most 5 minutes by default.
The maximum file size is 16MB by default.

Please don't upload anything illegal.

Inspiderd by [The Null Pointer](https://github.com/mia-0/0x0).
