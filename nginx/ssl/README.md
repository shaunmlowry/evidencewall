# SSL Certificates

Place your SSL certificates in this directory:

- `cert.pem` - SSL certificate
- `key.pem` - Private key

For development, you can generate self-signed certificates:

```bash
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes
```

For production, use certificates from a trusted CA like Let's Encrypt.
