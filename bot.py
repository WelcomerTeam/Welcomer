import discord

import config

client = discord.Client(intents=discord.Intents.all())

@client.event
async def on_member_join(member):
    await member.send("welcome!")

client.run(config.token)
