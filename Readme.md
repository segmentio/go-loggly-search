
# go-loggly-search

 Loggly search client for Go.

## Example

  Create a new search client with your account name and user credentials,
  then query against your logs with the fluent api:

```go
client := search.New("accountname", "username", "password")
res, err := client.Query(`(login OR logout) AND tobi`).Size(50).From("-5h").Fetch()
```

# License

MIT