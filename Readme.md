# Gilmo Backend

This is the backend that will support the Gilmo web app.
The backends role is to handle the interactions between the user and the gitlab instance. It will periodically check the jobs the user's runners are running as well as the runner's status

## Frameworks and services used

[Gin Gonic](github.com/gin-gonic/gin) - used to handle all http requests, authentication, logging, panic recovery, also using the sessions and cors plugins for Gin Gonic

[Nerzal's Gocloak](github.com/Nerzal/gocloak) - used for interacting with the Keycloak instance running behind the backend, handling user authentication

[Google's UUID](github.com/google/uuid) - used for unique identifiers

[Go-Gitlab from Xanzy](github.com/xanzy/go-gitlab) - used for interacting with the gitlab rest api

[Gorm ORM](gorm.io/gorm) - used for interacting with the mysql database

## Build instructions for local development

After cloning the repo, at the root, call "go mod tidy" to download all the dependencies than call "go build -o main.exe main.go" to build the project

## Api endpoints

/login - this is where user login data is sent and from where the access token is received from

/register - the api where the new users registration data is set. This api sends the data further to keycloak who handles the users

/authorized - this is base api endpoint for all other calls, all calls to this and child endpoints must come from a authorized user

/authorized/runners - this endpoint is used for getting the current user's runners, when calling with GET, or adding a new runner when calling with POST
/authorized/runners/{id} - this will return details about a runner when called with GET, edit a runner detail when using PUT/PATCH, delete a runner when called with DELETE

/authorized/runners/{id}/jobs - this endpoint returns a number of job objects for the specified runner when calling with GET

/authorized/statistics - this endpoint returns a number of statistics based on the input arguments. It displays the number of jobs, the average job duration, job failure rate and more
