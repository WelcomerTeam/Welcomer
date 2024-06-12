let EndpointGuild = function(guildID) {
  return `/api/guild/${guildID}`
}

let EndpointGuildAutorole = function(guildID) {
  return `${EndpointGuild(guildID)}/autoroles`;
}

let EndpointGuildBorderwall = function(guildID) {
  return `${EndpointGuild(guildID)}/borderwall`;
}

let EndpointGuildFreeroles = function(guildID) {
  return `${EndpointGuild(guildID)}/freeroles`;
}

let EndpointGuildLeaver = function(guildID) {
  return `${EndpointGuild(guildID)}/leaver`;
}

let EndpointGuildRules = function(guildID) {
  return `${EndpointGuild(guildID)}/rules`;
}

let EndpointGuildTempchannels = function(guildID) {
  return `${EndpointGuild(guildID)}/tempchannels`;
}

let EndpointGuildTimeroles = function(guildID) {
  return `${EndpointGuild(guildID)}/timeroles`;
}

let EndpointGuildWelcomer = function(guildID) {
  return `${EndpointGuild(guildID)}/welcomer`;
}

let EndpointGuildSettings = function(guildID) {
  return `${EndpointGuild(guildID)}/settings`;
}

export default {
  EndpointGuild,
  EndpointGuildAutorole,
  EndpointGuildBorderwall,
  EndpointGuildFreeroles,
  EndpointGuildLeaver,
  EndpointGuildRules,
  EndpointGuildTempchannels,
  EndpointGuildTimeroles,
  EndpointGuildWelcomer,
  EndpointGuildSettings
};
