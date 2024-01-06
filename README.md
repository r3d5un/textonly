# textonly

[![DigitalOcean Referral Badge](https://web-platforms.sfo2.cdn.digitaloceanspaces.com/WWW/Badge%201.svg)](https://www.digitalocean.com/?refcode=9934dd76e407&utm_campaign=Referral_Invite&utm_medium=Referral_Program&utm_source=badge)

Textonly is a simple blog app meant to serve as a reference web application written in Go, as well as my personal blog, which you can find [here](https://islandwind.me/).

The goal behind this project is to have a place where I can experiment with Go and other technologies. As such, the design is focused in exactly on what I want and need. Shortcuts have been made. Parts are under- and over-engineered, often at the same time. Other parts are just plain _bad_.

If you intend to use this project as a reference, make sure the design fits your problem and project.

## Applications

Textonly consists of two applications; the web application and a CLI for administering the web application.

I wanted an excuse to create a CLI tool, and have therefore limited the WebUI to be read-only. Any administration is intended to be done through the CLI.

## Todo

- [x] Basic WebUI
- [x] Web application configuration
- [x] API endpoints
- [x] API filters
- [x] API ordering
- [x] Secure API endpoints
- [ ] CLI
- [x] Dockerfile
- [ ] Documentation
- [ ] Swagger
