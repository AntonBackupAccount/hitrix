# Password
This service it can be used to hash and verify hashed passwords.

Register the service into your `main.go` file:
```go
 registry.ServiceProviderPassword(password.NewSimpleManager)
```

Access the service:
```go
service.DI().Password()
```