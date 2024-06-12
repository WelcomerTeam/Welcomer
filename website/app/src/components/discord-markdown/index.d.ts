interface DiscordCallback {
  user?: (id: number) => string;
  channel?: (id: number) => string;
  role?: (id: number) => string;
  everyone?: () => string;
  here?: () => string;
}

interface HTMLOptions {
  embed?: boolean;
  escapeHTML?: boolean;
  discordOnly?: boolean;
  discordCallback: DiscordCallback;
  cssModuleNames: Record<string, string>;
}

export function toHTML(source: string, options?: HTMLOptions): string;
