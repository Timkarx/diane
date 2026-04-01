I need to move out, but I can't be bothered to look through facebook posts for listings. Gonna try outsource this to Diane.


## WHY?

I tried openclaw, but it's completely fucked (not least because of typescript lol). I know what I need, so I'm just gonna make it. Also a good way to learn some Go.


## Architecture

I'm thinking the following:
- Agent (whatever that words means nowadays) at the core, essentially a wrapper around an LLM for injecting predefined prompts along with data
- Some mechanism for the main process to schedule it's own tasks (what is cron?) (i.e. telegram bot to pull latest listings from tg groups), then send it to the agent
- Some post-agent workflow to distribute the agent processing results (dm on telegram, send me an email...)
- Some sort of injest filtering/deduplication to not burn through $1,000,000 worth of tokens.
