# Stars

Compare stargazers across two projects on GitHub.

## Install

```sh
go get -u github.com/harshavardhana/github/stars
```

## Usage

Generate an `output.svg` a comparison chart for `minio/minio` v/s `mongodb/mongo`.
```sh
stars -mode=file -repos="minio/minio,mongodb/mongo" output.svg
```

With GitHub token
```
export GITHUB_TOKEN=xxxxxxxxxxx
stars -repo1 minio/minio -repo2 mongodb/mongo
```
