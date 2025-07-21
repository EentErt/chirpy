# Chirpy
## What is Chirpy?
Chirpy is a webtool for hosting a server platform where users can post short text posts of 140 characters. 
It includes authorization with JWT access tokens and refresh tokens and webhook access for including paid services.

## How to install
Chirpy requires a .env file which includes
```
DB_URL="[the url of a postgres database which stores user information and chirps]"
PLATFORM="[the host access level]"
SECRET="[the secret for encoding access tokens]"
POLKA_KEY="[the api key that polka will use to manage transaction processing]"
```
Chirpy requires a postgres database to store user information and chirp information.
