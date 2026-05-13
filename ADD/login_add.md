# Login Flow Design Document

## Overview

The tool requires a way for users to authenticate and authorize. This tool is indented to be used with gitlab so the most obvious solution is to integrate the auth process with it.

# Options available

For flexibility we could use our own keycloak integration and integrate it with gitlab.
Alternativelly we could remove the keycloak integration and just have directly implement user control and get the auth from gitlab

# Current implementation:
Integrates with keycloak for auth. Our service sends to keycloak user's credentials and receives a token, the service keeps and checks the token validity and refreshes it if necesary.

# Future implementation
Will only include user management without auth. When users login they are redirected to gitlab page where they enter their credentials. Upon successful login they are redirected to our service endpoint which stores the session