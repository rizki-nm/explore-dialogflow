# Explore Robolabs

This Go application serves as a webhook server for handling requests from Dialogflow, specifically designed to process intents.

## Overview

The server handles incoming POST requests from Dialogflow, processes the intent based on the request's session ID, retrieves relevant data from the Qiscus Omnichannel API, and formulates a response.

### Features

- **Webhook Handler**: Receives JSON payloads from Dialogflow containing intent details.
- **Qiscus API Integration**: Fetches room details and broadcast history from Qiscus Omnichannel API.
- **Response Processing**: Processes intent based on fetched data and constructs appropriate responses.

## Usage

### Run Locally

To run the project locally, follow these steps:

- Clone this repository and navigate to the directory
- Copy docker-compose.yml `cp docker-compose.yml.example docker-compose.yml`
- Start the App: `docker compose up`
