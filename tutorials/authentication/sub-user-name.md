# Sub-User Name Authentication

To create a new cloudns-go client with sub-user name authentication, you must
first collect the sub-user name and password from the ClouDNS control panel.

We will then pass the result of the `cloudns.AuthSubUserName()` function to the
`cloudns.New()` function.

```go
client, err := cloudns.New(
    cloudns.AuthSubUserName("john_doe", "JohnsPassword"),
)

if error != nil {
    fmt.Println("Error creating client:", err)
}
```

An if-statement similar to the one above can be used to check if a client was
successfully created.
