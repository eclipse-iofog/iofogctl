# Client Package

This package provides an HTTP client to communicate with ioFog Controller's REST API.

You can view see the full REST API specification at [iofog.org](https://iofog.org/docs/1.3.0/controllers/rest-api.html).

## Usage

First, instantiate a client instance and log in with your credentials.
```go
// Connect to Controller REST API
ctrl := client.New(endpoint)

// Create login request
loginRequest := client.LoginRequest{
	Email:    "user@domain.com",
	Password: "kj2gh0ooiwbug",
}

// Login
if err := ctrl.Login(loginRequest); err != nil {
	return err
}
```

Next, call any of the functions available from your client instance.
```go
// Get Controller status
if resp, err = ctrlClient.GetStatus(); err != nil {
    return err
}

// Print the response
println(resp.Status)
```