# gonetweet

Automatic tweet / X post destruction written in go

## Installation

```shell
go install github.com/junian/gonetweet@latest
```

Create `.env` file based on `.env.example`.

```shell
cp .env.example .env
```

Fill the `.env` content based on your Twitter / X API credentials.

## Examples

Use hashtag to schedule the deletion.

`d` for Days.

`h` for Hours.

This post will be deleted in 1 day:

```
I like fried chieckedn #1d
```

This post will be deleted in 1 day, 3 hours from the posted date:

```
Whatever doesn't kill you makes you stranger #1d3h
```

## License

This project is licensed under MIT License.

Made with ☕ by [Junian.dev](https://www.junian.dev).
