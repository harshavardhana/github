# Stars

Compare stargazers across two projects on GitHub.

## Install

```sh
go get -u github.com/harshavardhana/github/stars
```

## Usage

```sh
stars -repo1 minio/minio -repo2 mongodb/mongo
```

With GitHub token
```
export GITHUB_TOKEN=xxxxxxxxxxx
stars -repo1 minio/minio -repo2 mongodb/mongo
```

With Github token and custom page size.
```
export GITHUB_TOKEN=xxxxxxxxxxx
export GITHUB_PAGE_SIZE=1000
stars -repo1 minio/minio -repo2 mongodb/mongo
```
