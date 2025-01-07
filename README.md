# Description
***
This API provides a comprehensive solution for managing inventory borrowing processes with a strong focus on one-to-many relationships between the Transaction table and related tables. The API is built using the Gin framework and employs Gin_Auth for secure authentication and authorization.

# How to setup
***
1. create .env and insert
    * ```DATABASE_URL=your-username:your-password@tcp(localhost:3306)/your-db_name?charset=utf8mb4&parseTime=True&loc=Local```
    * ```JWT_SECRET_KEY = your-secret-key```
2. execute ```go mod init Gin-Inventory```
3. execute ```go mod tidy```
