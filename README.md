# deploy-test

## Usage

```bash
$ docker run -it --rm -v $(dirname $SSH_AUTH_SOCK):$(dirname $SSH_AUTH_SOCK) -e SSH_AUTH_SOCK=$SSH_AUTH_SOCK -p 3000:3000 deploy-test
```

## License

[GPLv3](COPYING)

## Author

[Benjamin Toll](https://benjamintoll.com)

