An autonomous background task agent (whatever that means nowadays).

## What?

A server that accepts data, runs into through an llm, executes various callbacks depending on what the llm decided to do with that information. Written in Go, and uses opencode under the hood.

## Why?

Because openclaw kinda sucks, and I have a good idea of what I need. 

## How?

Clone the project, define your own payload/notification types, and send data to the server. I'm working on cleaning up the api.

## Ideas
- Have your apps prouction errors summarized into possible causes so you can fix them quicker
- Scrape job postings and get notified when a new relevant position pops up (you have to build the scraper)
- Monitor a twitter account for mentions of certain topics (so you can beat everyone on polymarket)

## Lore
Named after Diane the personal assistant to special agent Dale Cooper from Twin Peaks.
