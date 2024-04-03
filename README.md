# serverless


## Installations

    Go (version 1.16 or newer) installed on your machine. Download Go.

## Steps to implement the project
    Create an environment variable

        Configure the following environment variables in a .env file or in your environment before running the application:

        DB_HOST

        DB_USER

        DB_PASSWORD

        DB_NAME (e.g., mywebappdb) - Ensure this database exists; the application will bootstrap or migrate the schema automatically.

        MAILGUN_API_KEY (for emailing)

        SENDER_DOMAIN (your domain)

        Clone the repository: git clone https://github.com/yourusername/yourrepositoryname.git cd serverless (Ensure that you have created the env file as it would have been ignored by git)

    Install the Go dependencies: go mod tidy

    Build the application: go build

    Run the application: go run main.go

## Structure of the cloud function (this project) project
    project-root-directory/
        go.mod
        email_function.go
