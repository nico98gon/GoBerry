# GoBerry

GoBerry is an open source user authentication and management system built with Go. It provides a robust and scalable solution for managing users and groups of users, featuring a double-factor authentication system and token-based authentication. GoBerry is designed to be integrated easily into any application via high-performance APIs, making it suitable for both small applications and those with millions of users.

## Features

**User Management**: CRUD operations for users.
**Group Management**: CRUD operations for groups of users.
**Authentication**: Double-factor authentication and token-based authentication.
**API**: High-performance REST APIs for user and group management.
**Extensibility**: Future plans include real-time chat and data analytics APIs.
**Scalability**: Designed to scale with Docker and Kubernetes.
## Getting Started

These instructions will help you get a copy of the project up and running on your local machine for development and testing purposes.

### Prerequisites

Go (version 1.22.4 or later)
PostgreSQL
Docker and Docker Compose
Git
### Installation

**Clone the repository**

```
git clone https://github.com/yourusername/GoBerry.git
cd GoBerry
```

**Set up the environment variables**

Create a .env file in the root directory and add your database URL:

```
DATABASE_URL=postgres://username
@localhost:5432/mydb?sslmode=disable
```

**Install dependencies**

```
go mod download
```

**Run the database migrations**

Create a migrations directory inside db and add your SQL scripts. For now, you can manually create the users table:

```
CREATE TABLE IF NOT EXISTS users (
id SERIAL PRIMARY KEY,
name TEXT,
email TEXT
);
```

**Build and run the application**

```
go build -o bin/go-berry
./bin/go-berry
```

### Using Docker

**Build the Docker image**

```
docker build -t goberry .
```

**Run the Docker container**

```
docker run -p 8080:8080 --env-file .env goberry
```

### Using Docker Compose

**Run the application**

```
docker-compose up
```

## API Endpoints

**GET /users**: Retrieve all users
**GET /users/{id}**: Retrieve a user by ID
**POST /users**: Create a new user
**PUT /users/{id}**: Update a user by ID
**DELETE /users/{id}**: Delete a user by ID
## Contributing

We welcome contributions to GoBerry! Please fork the repository and create a pull request with your changes. For major changes, please open an issue first to discuss what you would like to change.

Fork the Project
Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
Push to the Branch (`git push origin feature/AmazingFeature`)
Open a Pull Request
## License

Distributed under the MIT License. See `LICENSE` for more information.

## Acknowledgements

Gorilla Mux for HTTP routing and URL matching.
GORM for ORM.
PostgreSQL for the database.
Docker and Kubernetes for containerization and orchestration.