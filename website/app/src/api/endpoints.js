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

let EndpointGuildWelcomerBuilder = function(guildID) {
  return `${EndpointGuildWelcomer(guildID)}/builder`;
}

let EndpointGuildWelcomerBuilderPreview = function(guildID) {
  return `${EndpointGuildWelcomerBuilder(guildID)}/preview`;
}

let EndpointGuildWelcomerBuilderArtifact = function(guildID) {
  return `${EndpointGuildWelcomer(guildID)}/artifact`;
}

let EndpointGuildWelcomerViewBuilderArtifact = function(guildID, artifactRef) {
  return `${EndpointGuildWelcomerBuilderArtifact(guildID)}/${artifactRef}`;
}

let EndpointGuildSettings = function(guildID) {
  return `${EndpointGuild(guildID)}/settings`;
}

let EndpointGuildSettingsUpdateMemberCount = function(guildID) {
  return `${EndpointGuildSettings(guildID)}/update-member-count`;
}

let EndpointGuildCustomisation = function(guildID) {
  return `${EndpointGuild(guildID)}/customisation`;
}

let EndpointGuildCustomBots = function(guildID) {
  return `${EndpointGuild(guildID)}/custom-bots`;
}

let EndpointGuildCustomBot = function(guildID, botID) {
  return `${EndpointGuildCustomBots(guildID)}/${botID}`;
}

let EndpointStartGuildCustomBot = function(guildID, botID) {
  return `${EndpointGuildCustomBot(guildID, botID)}/start`;
}

let EndpointStopGuildCustomBot = function(guildID, botID) {
  return `${EndpointGuildCustomBot(guildID, botID)}/stop`;
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
  EndpointGuildWelcomerBuilder,
  EndpointGuildWelcomerBuilderArtifact,
  EndpointGuildWelcomerViewBuilderArtifact,
  EndpointGuildWelcomerBuilderPreview,
  EndpointGuildSettings,
  EndpointGuildSettingsUpdateMemberCount,
  EndpointGuildCustomisation,
  EndpointGuildCustomBots,
  EndpointGuildCustomBot,
  EndpointStartGuildCustomBot,
  EndpointStopGuildCustomBot
};
