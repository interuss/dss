variable "authorization" {
  type = object({
    public_key_pem_path = optional(string)
    jwks = optional(object({
      endpoint = string
      key_id   = string
    }))
  })
  description = <<EOT
    One of `public_key_pem_path` or `jwks` should be provided but not both.

    - public_key_pem_path
      If providing the access token public key via JWKS, do not provide this parameter.
      If providing a .pem file directly as the public key to validate incoming access tokens, specify the name
      of this .pem file here as /public-certs/YOUR-KEY-NAME.pem replacing YOUR-KEY-NAME as appropriate. For instance,
      if using the provided us-demo.pem, use the path /public-certs/us-demo.pem. Note that your .pem file should built
      in the docker image or mounted manually.

      Example 1 (dummy auth):
      ```
      {
        public_key_pem_path = "/test-certs/auth2.pem"
      }
      ```
      Example 2:
      ```
      {
        public_key_pem_path = "/jwt-public-certs/us-demo.pem"
      }
      ```

    - jwks
        If providing a .pem file directly as the public key to validate incoming access tokens, do not provide this parameter.
        - endpoint
          If providing the access token public key via JWKS, specify the JWKS endpoint here.
          Example: https://auth.example.com/.well-known/jwks.json
        - key_id:
          If providing the access token public key via JWKS, specify the kid (key ID) of they appropriate key in the JWKS file referenced above.
      Example:
      ```
      {
        jwks = {
          endpoint = "https://auth.example.com/.well-known/jwks.json"
          key_id = "9C6DF78B-77A7-4E89-8990-E654841A7826"
        }
      }
      ```
  EOT

  validation {
    condition     = (var.authorization.jwks == null && var.authorization.public_key_pem_path != null) || (var.authorization.jwks != null && var.authorization.public_key_pem_path == null)
    error_message = "Public key to validate incoming access tokens shall be provided exclusively either with a .pem file or via JWKS."
  }
}