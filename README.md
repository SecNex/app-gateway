# SecNex Application Gateway

The SecNex Application Gateway is a reverse proxy that provides security features to protect web applications. It is designed to be a lightweight and easy to use solution for securing web applications.

## Getting Started

### Prerequisites

- Docker
- Docker Compose

### Installation

1. Clone the repository
    ```bash
    git clone https://scripts.secnex.io/secnex-application-gateway/v1.git
    cd secnex-application-gateway
    ```

2. Create a `.env` file
    ```bash
    cp .env.example .env
    ```

3. Start the application
    ```bash
    docker-compose up -d
    ```

4. Access the admin panel at `http://localhost:8080`

## Features

- Authentication
- Web Application Firewall (WAF)
    - Rate Limiting
    - IP Whitelisting / Blacklisting
    - User-Agent Whitelisting / Blacklisting
    - Path Whitelisting / Blacklisting
    - SQL Injection Protection
    - Method Whitelisting / Blacklisting - can set for each path or globally

### Authentication

We use a postgres database to store API keys and their associated users. The gateway will check the API key in the request header and authenticate the user based on the key.

You can integrate your own identity provider or use the built-in database.

### Web Application Firewall (WAF)

The WAF provides a set of security features to protect your web application from common attacks. You can configure the WAF to protect your application based on your requirements.