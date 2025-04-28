<template>
  <div class="message-2qnXI6 cozyMessage-3V1Y8y groupStart-23k01U wrapper-2a6GCs cozy-3raOZG zalgo-jN1Ica"
    role="listitem">
    <div class="contents-2mQqc9" role="document">
      <img :src="avatar" aria-hidden="true" class="avatar-1BDn8e" alt="Message author icon" />
      <h2 class="header-23xsNx">
        <span class="headerText-3Uvj1Y"><span :class="[
          $props.isDark
            ? 'text-gray-50'
            : ('text-secondary ' + ($props.respectDarkMode ? 'dark:text-gray-50' : '')),
          'username-1A8OIy desaturateUserColors-1gar-1',
        ]" aria-expanded="false" tabindex="0" :style="authorColour ? { color: `#${rgbIntToRGB(authorColour)}` } : {}
          ">
            {{ formatText(author) }}</span><span v-if="isBot"
            class="botTagCozy-1fFsZk botTag-1un5a6 botTagRegular-2HEhHi botTag-2WPJ74 rem-2m9HGf"><svg
              aria-label="Verified Bot" class="botTagVerified-1klIIt" aria-hidden="false" width="16" height="16"
              viewBox="0 0 16 15.2">
              <path d="M7.4,11.17,4,8.62,5,7.26l2,1.53L10.64,4l1.36,1Z" fill="currentColor"></path>
            </svg><span class="botText-1526X_">BOT</span></span></span><span
          class="timestamp-3ZCmNB timestampInline-yHQ6fX" v-if="showTimestamp"><time :aria-label="timestamp"
            :datetime="now"><i class="separator-2nZzUB" aria-hidden="true"> — </i>{{ formatText(timestamp) }}</time></span>
      </h2>
      <div :class="[
        $props.isDark ? 'text-gray-50' : 'text-secondary dark:text-gray-50',
        'markup-2BOw-j messageContent-2qWWxC',
      ]" v-html="marked(content, true)" />
    </div>
    <div class="container-1ov-mD" v-for="embed in embeds" v-bind:key="embed">
      <div class="embedWrapper-lXpS3L embedFull-2tM8-- embed-IeVjo6 markup-2BOw-j" aria-hidden="false"
        :style="{ 'border-color': `${rgbIntToRGB(embed?.color, 2450411)}` }">
        <div :class="[
          'grid-1nZz7S',
          embed?.thumbnail?.url ? 'hasThumbnail-3FJf1w' : '',
        ]">
          <div class="embedAuthor-3l5luH embedMargin-UO5XwE" v-if="embed?.author">
            <img aria-hidden="true" alt="Embed author icon" class="embedAuthorIcon--1zR3L"
              :src="formatText(embed?.author?.icon_url)" v-if="embed?.author?.icon_url" /><a v-if="embed?.author?.url"
              class="anchor-3Z-8Bb anchorUnderlineOnHover-2ESHQB embedAuthorNameLink-1gVryT embedLink-1G1K1D embedAuthorName-3mnTWj"
              tabindex="0" href="#" rel="noreferrer noopener">{{ formatText(embed?.author?.name) }}</a>
            <span v-else class="embedAuthorName-3mnTWj">
              {{ formatText(embed?.author?.name) }}
            </span>
          </div>
          <div class="embedTitle-3OXDkz embedMargin-UO5XwE" v-if="embed?.title">
            <a class="anchor-3Z-8Bb anchorUnderlineOnHover-2ESHQB embedTitleLink-1Zla9e embedLink-1G1K1D embedTitle-3OXDkz"
              tabindex="0" :href="enableURLs ? embed?.url : '#'" rel="noreferrer noopener"
              :role="embed?.url ? 'button' : ''" v-html="marked(embed?.title, true)" />
          </div>
          <div class="embedDescription-1Cuq9a embedMargin-UO5XwE" v-html="marked(embed?.description, true)"
            v-if="embed?.description" />
          <div class="embedFields-2IPs5Z">
            <div class="embedField-1v-Pnh" :style="'grid-column: ' +
              (field.inline ? (field.odd ? '7 / 13' : '1 / 7') : '1 / 13')
              " v-for="field in embed?.fields" v-bind:key="field">
              <div class="embedFieldName-NFrena">
                <span class="emojiContainer-3X8SvE" tabindex="0" v-html="marked(field.name, true)" />
              </div>
              <div class="embedFieldValue-nELq2s">
                <span v-html="marked(field.value, true)" />
              </div>
            </div>
          </div>
          <div class="anchor-3Z-8Bb anchorUnderlineOnHover-2ESHQB imageWrapper-2p5ogY clickable-3Ya1ho embedWrapper-lXpS3L embedMedia-1guQoW embedImage-2W1cML"
            tabindex="0" href="#" rel="noreferrer noopener" v-if="embed?.image?.url"><img
              aria-hidden="true" alt="Embed image" :src="formatText(embed?.image?.url)" /></div>
          <div class="anchor-3Z-8Bb anchorUnderlineOnHover-2ESHQB imageWrapper-2p5ogY clickable-3Ya1ho embedThumbnail-2Y84-K"
            tabindex="0" href="#" rel="noreferrer noopener" style="width: 80px; height: 80px"
            v-if="embed?.thumbnail?.url"><img aria-hidden="true" alt="Embed thumbnail" :src="formatText(embed?.thumbnail?.url)"
              style="width: 80px; height: 80px" /></div>
          <div class="embedFooter-3yVop- embedMargin-UO5XwE">
            <img class="embedFooterIcon-3klTIQ" :src="formatText(embed?.footer?.icon_url)" v-if="embed?.footer?.icon_url" />
            <span class="embedFooterText-28V_Wb">{{ formatText(embed?.footer?.text) }}<span class="embedFooterSeparator-3klTIQ" v-if="embed?.footer?.text && showTimestamp">•</span><span
                v-if="showTimestamp">{{ formatText(timestamp) }}</span></span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="css">
.cozy-3raOZG .headerText-3Uvj1Y {
  margin-right: 0.25rem;
}

.botTagCozy-1fFsZk {
  margin-left: 0.25rem;
}

.embedAuthorIcon--1zR3L {
  margin-right: 8px;
  width: 24px;
  height: 24px;
  object-fit: contain;
  border-radius: 50%;
}

.rem-2m9HGf.botTag-2WPJ74 {
  height: 0.9375rem;
  padding: 0 0.275rem;
  margin-top: 0.075em;
  border-radius: 0.1875rem;
}

.rem-2m9HGf .botTagVerified-1klIIt {
  width: 0.9375rem;
  height: 0.9375rem;
  margin-left: -0.25rem;
}

.rem-2m9HGf .botText-1526X_ {
  line-height: 0.9375rem;
}

.wrapper-2a6GCs {
  position: relative;
  overflow-wrap: break-word;
  user-select: text;
  -webkit-box-flex: 0;
  flex: 0 0 auto;
  padding-right: 16px;
  min-height: 1.375rem;
}

.cozy-3raOZG.wrapper-2a6GCs {
  padding-top: 0.125rem;
  padding-bottom: 0.125rem;
}

.cozy-3raOZG.wrapper-2a6GCs {
  padding-left: 72px;
}

.botTag-1un5a6 {
  position: relative;
  top: 0.1rem;
}

.embedImage-2W1cML,
.embedThumbnail-2Y84-K {
  display: block;
  object-fit: fill;
}

.embedImage-2W1cML img,
.embedThumbnail-2Y84-K img {
  display: block;
  border-radius: 4px;
}

.embedThumbnail-2Y84-K {
  grid-area: 1/2/8/2;
  margin-left: 16px;
  margin-top: 8px;
  flex-shrink: 0;
  justify-self: end;
}

.botTag-2WPJ74 {
  font-size: 0.625rem;
  text-transform: uppercase;
  vertical-align: top;
  display: inline-flex;
  -webkit-box-align: center;
  align-items: center;
  flex-shrink: 0;
  text-indent: 0;
}

.rem-2m9HGf.botTag-2WPJ74 {
  height: 0.9375rem;
  padding: 0 0.275rem;
  margin-top: 0.075em;
  border-radius: 0.1875rem;
}

.cozy-3raOZG.wrapper-2a6GCs {
  padding-top: 0.125rem;
  padding-bottom: 0.125rem;
}

.cozy-3raOZG.wrapper-2a6GCs {
  padding-left: 72px;
}

.cozy-3raOZG .contents-2mQqc9 {
  position: static;
  margin-left: 0;
  padding-left: 0;
  text-indent: 0;
}

.cozy-3raOZG .header-23xsNx {
  display: block;
  position: relative;
  line-height: 1.375rem;
  min-height: 1.375rem;
  color: var(--text-muted);
  white-space: break-spaces;
}

.zalgo-jN1Ica.cozy-3raOZG .header-23xsNx {
  overflow: hidden;
}

.cozy-3raOZG .timestamp-3ZCmNB {
  font-size: 0.75rem;
  line-height: 1.375rem;
  color: var(--text-muted);
  vertical-align: baseline;
}

.cozy-3raOZG .headerText-3Uvj1Y {
  margin-right: 0.25rem;
}

.cozy-3raOZG .messageContent-2qWWxC {
  position: relative;
}

.cozy-3raOZG .messageContent-2qWWxC {
  user-select: text;
  margin-left: -72px;
  padding-left: 72px;
}

.markup-2BOw-j {
  font-size: 1rem;
  line-height: 1.375rem;
  white-space: break-spaces;
  overflow-wrap: break-word;
  user-select: text;
  color: var(--text-normal);
  font-weight: 400;
}

.markup-2BOw-j a {
  color: var(--text-link);
  word-break: break-word;
  text-decoration: none;
  cursor: pointer;
}

.markup-2BOw-j a:hover {
  text-decoration: underline;
}

.markup-2BOw-j pre {
  border-radius: 4px;
  padding: 0;
  font-family: Consolas, "Andale Mono WT", "Andale Mono", "Lucida Console",
    "Lucida Sans Typewriter", "DejaVu Sans Mono", "Bitstream Vera Sans Mono",
    "Liberation Mono", "Nimbus Mono L", Monaco, "Courier New", Courier,
    monospace;
  font-size: 0.75rem;
  line-height: 1rem;
  margin-top: 6px;
  white-space: pre-wrap;
  background-clip: border-box;
}

.markup-2BOw-j pre {
  box-sizing: border-box;
  max-width: 90%;
}

.markup-2BOw-j code {
  font-size: 0.875rem;
  line-height: 1.125rem;
  text-indent: 0;
  white-space: pre-wrap;
  background: var(--background-secondary);
  border: 1px solid var(--background-tertiary);
  color: var(--text-normal);
}

.markup-2BOw-j code.inline {
  width: auto;
  height: auto;
  padding: 0.2em;
  margin: -0.2em 0;
  border-radius: 3px;
  font-size: 85%;
  font-family: Consolas, "Andale Mono WT", "Andale Mono", "Lucida Console",
    "Lucida Sans Typewriter", "DejaVu Sans Mono", "Bitstream Vera Sans Mono",
    "Liberation Mono", "Nimbus Mono L", Monaco, "Courier New", Courier,
    monospace;
  text-indent: 0;
  border: none;
  white-space: pre-wrap;
}

.embedAuthorName-3mnTWj,
.embedAuthorNameLink-1gVryT,
.embedDescription-1Cuq9a,
.embedFieldName-NFrena,
.embedFieldValue-nELq2s,
.embedFooterText-28V_Wb,
.embedLink-1G1K1D,
.embedTitle-3OXDkz,
.embedTitleLink-1Zla9e {
  unicode-bidi: plaintext;
  text-align: left;
}

.embedFooterIcon-3klTIQ {
  margin-right: 8px;
  width: 20px;
  height: 20px;
  -o-object-fit: contain;
  object-fit: contain;
  border-radius: 50%;
}

.embedFieldValue-nELq2s {
  font-size: 0.875rem;
  line-height: 1.125rem;
  font-weight: 400;
  white-space: pre-line;
  min-width: 0;
}

.embedDescription-1Cuq9a,
.embedFieldValue-nELq2s {
  color: var(--text-normal);
}

.wrapper-2aW0bm {
  background-color: var(--background-primary);
  box-shadow: var(--elevation-stroke);
  display: grid;
  grid-auto-flow: column;
  box-sizing: border-box;
  height: 32px;
  border-radius: 4px;
  -webkit-box-align: center;
  align-items: center;
  -webkit-box-pack: start;
  justify-content: flex-start;
  user-select: none;
  transition: box-shadow 0.1s ease-out 0s, -webkit-box-shadow 0.1s ease-out 0s;
  position: relative;
  overflow: hidden;
}

.cozy-3raOZG .header-23xsNx {
  display: block;
  position: relative;
  line-height: 1.375rem;
  min-height: 1.375rem;
  color: var(--text-muted);
  white-space: break-spaces;
}

.zalgo-jN1Ica.cozy-3raOZG .header-23xsNx {
  overflow: hidden;
}

.zalgo-jN1Ica.cozy-3raOZG .header-23xsNx {
  overflow: hidden;
}

.zalgo-jN1Ica .messageContent-2qWWxC {
  overflow: hidden;
}

.embedAuthorName-3mnTWj,
.embedAuthorNameLink-1gVryT,
.embedDescription-1Cuq9a,
.embedFieldName-NFrena,
.embedFieldValue-nELq2s,
.embedFooterText-28V_Wb,
.embedLink-1G1K1D,
.embedTitle-3OXDkz,
.embedTitleLink-1Zla9e {
  unicode-bidi: plaintext;
  text-align: left;
}

.embedField-1v-Pnh,
.embedFieldName-NFrena {
  font-size: 0.875rem;
  line-height: 1.125rem;
  min-width: 0;
}

.embedFieldName-NFrena {
  font-weight: 600;
  margin-bottom: 2px;
}

.embedAuthorName-3mnTWj,
.embedFieldName-NFrena,
.embedTitle-3OXDkz {
  color: var(--header-primary);
}

.embedAuthorName-3mnTWj,
.embedAuthorNameLink-1gVryT,
.embedDescription-1Cuq9a,
.embedFieldName-NFrena,
.embedFieldValue-nELq2s,
.embedFooterText-28V_Wb,
.embedLink-1G1K1D,
.embedTitle-3OXDkz,
.embedTitleLink-1Zla9e {
  unicode-bidi: plaintext;
  text-align: left;
}

.embedFooterText-28V_Wb {
  font-size: 0.75rem;
  line-height: 1rem;
  font-weight: 500;
  color: var(--text-normal);
}

.avatar-1BDn8e.clickable-1bVtEA {
  pointer-events: auto;
}

.avatar-1BDn8e.clickable-1bVtEA:active {
  transform: translateY(1px);
}

.emojiContainer-3X8SvE {
  display: inline-block;
}

.anchorUnderlineOnHover-2ESHQB:hover {
  text-decoration: underline;
}

.embedImage-2W1cML,
.embedThumbnail-2Y84-K {
  display: block;
  object-fit: fill;
}

.embedImage-2W1cML img,
.embedThumbnail-2Y84-K img {
  display: block;
  border-radius: 4px;
}

.embed-IeVjo6 {
  position: relative;
  display: grid;
  max-width: 520px;
  box-sizing: border-box;
  border-radius: 4px;
}

.embed-IeVjo6 pre {
  max-width: 100%;
  border: none;
}

.embed-IeVjo6 code {
  border: none;
  background: var(--background-tertiary);
}

.embed-IeVjo6 .embedAuthorNameLink-1gVryT {
  color: var(--header-primary);
}

.grid-1nZz7S {
  overflow: hidden;
  padding: 0.5rem 1rem 1rem 0.75rem;
  display: inline-grid;
  grid-template-columns: auto;
  grid-template-rows: auto;
}

.grid-1nZz7S.hasThumbnail-3FJf1w {
  grid-template-columns: auto min-content;
}

.embedMedia-1guQoW {
  grid-column: 1/1;
  border-radius: 4px;
  contain: paint;
}

.hasThumbnail-3FJf1w .embedMedia-1guQoW {
  grid-column: 1/3;
}

.embedFull-2tM8-- .embedMedia-1guQoW {
  margin-top: 16px;
}

.embedAuthor-3l5luH,
.embedDescription-1Cuq9a,
.embedFields-2IPs5Z,
.embedFooter-3yVop-,
.embedMedia-1guQoW,
.embedTitle-3OXDkz {
  min-width: 0;
}

.imageWrapper-2p5ogY {
  display: block;
  position: relative;
  user-select: text;
  overflow: hidden;
  border-radius: 3px;
}

.message-2qnXI6 {
  padding-right: 48px !important;
}

.botTagVerified-1klIIt {
  display: inline-block;
}

.rem-2m9HGf .botTagVerified-1klIIt {
  width: 0.9375rem;
  height: 0.9375rem;
  margin-left: -0.25rem;
}

.messageContent-2qWWxC {
  text-indent: 0;
}

.cozy-3raOZG .messageContent-2qWWxC {
  position: relative;
}

.zalgo-jN1Ica .messageContent-2qWWxC {
  overflow: hidden;
}

.messageContent-2qWWxC:empty {
  display: none;
}

.cozy-3raOZG .messageContent-2qWWxC {
  user-select: text;
  margin-left: -72px;
  padding-left: 72px;
}

.anchor-3Z-8Bb {
  color: var(--text-link);
  text-decoration: none;
}

.avatar-1BDn8e {
  position: absolute;
  left: 16px;
  margin-top: calc(4px - 0.125rem);
  width: 40px;
  height: 40px;
  border-radius: 50%;
  overflow: hidden;
  user-select: none;
  -webkit-box-flex: 0;
  flex: 0 0 auto;
  pointer-events: none;
  z-index: 1;
}

.avatar-1BDn8e.clickable-1bVtEA {
  pointer-events: auto;
}

.avatar-1BDn8e.clickable-1bVtEA:active {
  transform: translateY(1px);
}

.embedAuthorName-3mnTWj,
.embedAuthorNameLink-1gVryT,
.embedDescription-1Cuq9a,
.embedFieldName-NFrena,
.embedFieldValue-nELq2s,
.embedFooterText-28V_Wb,
.embedLink-1G1K1D,
.embedTitle-3OXDkz,
.embedTitleLink-1Zla9e {
  unicode-bidi: plaintext;
  text-align: left;
}

.embedAuthorName-3mnTWj {
  font-size: 0.875rem;
  font-weight: 600;
}

.embedAuthorName-3mnTWj,
.embedFieldName-NFrena,
.embedTitle-3OXDkz {
  color: var(--header-primary);
}

.emoji {
  object-fit: contain;
  width: 1.375em;
  height: 1.375em;
  vertical-align: bottom;
  display: inline;
}

.emojiContainer-3X8SvE {
  display: inline-block;
}

.embed-IeVjo6 .emoji {
  width: 18px;
  height: 18px;
  display: inline;
}

.emoji-3C344l {
  margin-right: 8px;
  object-fit: contain;
  background-position: 50% center;
  background-repeat: no-repeat;
}

.emoji-3C344l {
  margin-bottom: 8px;
}

.emoji-270c6v {
  margin-right: 8px;
}

.embedFooter-3yVop- {
  display: flex;
  -webkit-box-align: center;
  align-items: center;
  grid-area: auto/1/auto/1;
}

.hasThumbnail-3FJf1w .embedFooter-3yVop- {
  grid-column: 1/3;
}

.embedAuthor-3l5luH,
.embedDescription-1Cuq9a,
.embedFields-2IPs5Z,
.embedFooter-3yVop-,
.embedMedia-1guQoW,
.embedTitle-3OXDkz {
  min-width: 0;
}

.scrollbar-3dvm_9::-webkit-scrollbar-corner {
  border: none;
  background: 0 0;
}

.hljs {
  display: block;
  overflow-x: auto;
  padding: 0.5em;
  border-radius: 4px;
  color: var(--header-secondary);
  text-size-adjust: none;
}

.hljs-name,
.hljs-title {
  color: #268bd2;
}

.hljs-class .hljs-title {
  color: #b58900;
}

.hljs {
  display: block;
  overflow-x: auto;
  padding: 0.5em;
  border-radius: 4px;
  color: var(--header-secondary);
  text-size-adjust: none;
}

.hljs-name,
.hljs-title {
  color: #268bd2;
}

.hljs-class .hljs-title {
  color: #b58900;
}

.botText-1526X_ {
  position: relative;
  font-weight: 500;
}

.rem-2m9HGf .botText-1526X_ {
  line-height: 0.9375rem;
}

.isHeader-2dII4U {
  top: -16px;
}

.username-1A8OIy {
  font-size: 1rem;
  font-weight: 500;
  line-height: 1.375rem;
  color: var(--header-primary);
  display: inline;
  vertical-align: baseline;
  position: relative;
  overflow: hidden;
  flex-shrink: 0;
}

.username-1A8OIy {
  pointer-events: none;
}

.cozy-3raOZG .contents-2mQqc9 {
  position: static;
  margin-left: 0;
  padding-left: 0;
  text-indent: 0;
}

.timestampInline-yHQ6fX {
  margin-left: 0.25rem;
}

.botTagRegular-2HEhHi {
  background: var(--brand-experiment);
  color: #fff;
}

.separator-2nZzUB {
  position: absolute;
  opacity: 0;
  width: 0;
  display: inline-block;
  font-style: normal;
}

.scrollbarGhostHairline-1mSOM1::-webkit-scrollbar {
  width: 4px;
  height: 4px;
}

.scrollbarGhostHairline-1mSOM1::-webkit-scrollbar-thumb {
  background-color: rgba(24, 25, 28, 0.6);
  border-radius: 2px;
  cursor: move;
}

.scrollbarGhostHairline-1mSOM1::-webkit-scrollbar-track {
  background-color: transparent;
  border: none;
}

.embedFooterSeparator-3klTIQ {
  font-weight: 500;
  color: var(--text-normal);
  display: inline-block;
  margin: 0 4px;
}

.button-1ZiXG9 {
  display: flex;
  -webkit-box-align: center;
  align-items: center;
  -webkit-box-pack: center;
  justify-content: center;
  height: 24px;
  padding: 4px;
  min-width: 24px;
  -webkit-box-flex: 0;
  flex: 0 0 auto;
  color: var(--interactive-normal);
  cursor: pointer;
  position: relative;
}

.button-1ZiXG9:hover {
  color: var(--interactive-hover);
  background-color: var(--background-modifier-hover);
}

.button-1ZiXG9:active {
  padding-top: 5px;
  padding-bottom: 3px;
  color: var(--interactive-active);
  background-color: var(--background-modifier-active);
}

.icon-3Gkjwa {
  width: 20px;
  height: 20px;
  display: block;
  object-fit: contain;
}

.embedAuthorName-3mnTWj,
.embedAuthorNameLink-1gVryT,
.embedDescription-1Cuq9a,
.embedFieldName-NFrena,
.embedFieldValue-nELq2s,
.embedFooterText-28V_Wb,
.embedLink-1G1K1D,
.embedTitle-3OXDkz,
.embedTitleLink-1Zla9e {
  unicode-bidi: plaintext;
  text-align: left;
}

.embed-IeVjo6 .embedAuthorNameLink-1gVryT {
  color: var(--header-primary);
}

.buttonContainer-DHceWr {
  position: absolute;
  top: 0;
  right: 0;
}

.embedAuthorName-3mnTWj,
.embedAuthorNameLink-1gVryT,
.embedDescription-1Cuq9a,
.embedFieldName-NFrena,
.embedFieldValue-nELq2s,
.embedFooterText-28V_Wb,
.embedLink-1G1K1D,
.embedTitle-3OXDkz,
.embedTitleLink-1Zla9e {
  unicode-bidi: plaintext;
  text-align: left;
}

.embedDescription-1Cuq9a {
  font-size: 0.875rem;
  line-height: 1.125rem;
  font-weight: 400;
  white-space: pre-line;
  grid-column: 1/1;
}

.embedDescription-1Cuq9a,
.embedFieldValue-nELq2s {
  color: var(--text-normal);
}

.embedAuthor-3l5luH,
.embedDescription-1Cuq9a,
.embedFields-2IPs5Z,
.embedFooter-3yVop-,
.embedMedia-1guQoW,
.embedTitle-3OXDkz {
  min-width: 0;
}

.grid-1nZz7S.hasThumbnail-3FJf1w {
  grid-template-columns: auto min-content;
}

.hasThumbnail-3FJf1w .embedFooter-3yVop- {
  grid-column: 1/3;
}

.hasThumbnail-3FJf1w .embedMedia-1guQoW {
  grid-column: 1/3;
}

.embedAuthorName-3mnTWj,
.embedAuthorNameLink-1gVryT,
.embedDescription-1Cuq9a,
.embedFieldName-NFrena,
.embedFieldValue-nELq2s,
.embedFooterText-28V_Wb,
.embedLink-1G1K1D,
.embedTitle-3OXDkz,
.embedTitleLink-1Zla9e {
  unicode-bidi: plaintext;
  text-align: left;
}

.embedTitle-3OXDkz {
  font-size: 1rem;
  font-weight: 600;
  display: inline-block;
  grid-column: 1/1;
}

.embedAuthorName-3mnTWj,
.embedFieldName-NFrena,
.embedTitle-3OXDkz {
  color: var(--header-primary);
}

.embedAuthor-3l5luH,
.embedDescription-1Cuq9a,
.embedFields-2IPs5Z,
.embedFooter-3yVop-,
.embedMedia-1guQoW,
.embedTitle-3OXDkz {
  min-width: 0;
}

.embedField-1v-Pnh {
  font-weight: 400;
}

.embedField-1v-Pnh,
.embedFieldName-NFrena {
  font-size: 0.875rem;
  line-height: 1.125rem;
  min-width: 0;
}

.buttons-cl5qTG {
  opacity: 0;
  pointer-events: none;
}

.embedFields-2IPs5Z {
  display: grid;
  grid-column: 1/1;
  margin-top: 8px;
  gap: 8px;
}

.embedAuthor-3l5luH,
.embedDescription-1Cuq9a,
.embedFields-2IPs5Z,
.embedFooter-3yVop-,
.embedMedia-1guQoW,
.embedTitle-3OXDkz {
  min-width: 0;
}

.timestamp-3ZCmNB {
  display: inline-block;
  height: 1.25rem;
  cursor: default;
  pointer-events: none;
  font-weight: 500;
}

.cozy-3raOZG .timestamp-3ZCmNB {
  font-size: 0.75rem;
  line-height: 1.375rem;
  color: var(--text-muted);
  vertical-align: baseline;
}

.embedAuthor-3l5luH {
  display: flex;
  -webkit-box-align: center;
  align-items: center;
  grid-column: 1/1;
}

.embedAuthor-3l5luH,
.embedDescription-1Cuq9a,
.embedFields-2IPs5Z,
.embedFooter-3yVop-,
.embedMedia-1guQoW,
.embedTitle-3OXDkz {
  min-width: 0;
}

.cozyMessage-3V1Y8y.groupStart-23k01U {
  min-height: 2.75rem;
}

.embedFull-2tM8-- {
  border-left: 4px solid var(--background-tertiary);
  background: var(--background-secondary);
}

.embedFull-2tM8-- .embedMedia-1guQoW {
  margin-top: 16px;
}

.cozyMessage-3V1Y8y.groupStart-23k01U {
  min-height: 2.75rem;
}

.embedAuthorName-3mnTWj,
.embedAuthorNameLink-1gVryT,
.embedDescription-1Cuq9a,
.embedFieldName-NFrena,
.embedFieldValue-nELq2s,
.embedFooterText-28V_Wb,
.embedLink-1G1K1D,
.embedTitle-3OXDkz,
.embedTitleLink-1Zla9e {
  unicode-bidi: plaintext;
  text-align: left;
}

.embedMargin-UO5XwE {
  margin-top: 8px;
}

.container-1ov-mD {
  display: grid;
  grid-auto-flow: row;
  row-gap: 0.25rem;
  text-indent: 0;
  min-height: 0;
  min-width: 0;
  padding-top: 0.125rem;
  padding-bottom: 0.125rem;
  position: relative;
}

.container-1ov-mD:empty {
  display: none;
}

.container-1ov-mD>* {
  place-self: start;
}

.container-3npvBV {
  position: absolute;
  right: 0;
  z-index: 1;
  top: -25px;
  padding: 0 14px 0 32px;
}

.embedAuthorName-3mnTWj,
.embedAuthorNameLink-1gVryT,
.embedDescription-1Cuq9a,
.embedFieldName-NFrena,
.embedFieldValue-nELq2s,
.embedFooterText-28V_Wb,
.embedLink-1G1K1D,
.embedTitle-3OXDkz,
.embedTitleLink-1Zla9e {
  unicode-bidi: plaintext;
  text-align: left;
}

.embedLink-1G1K1D {
  text-decoration: none;
  cursor: pointer;
}

.embedLink-1G1K1D:hover {
  text-decoration: underline;
}

* {
  --header-primary: #fff;
  --header-secondary: #b9bbbe;
  --text-normal: #dcddde;
  --text-muted: #72767d;
  --text-link: hsl(197, 100, 47.8%);
  --background-primary: #36393f;
  --background-secondary: #2f3136;
  --background-secondary-alt: #292b2f;
  --background-tertiary: #202225;
  --background-accent: #4f545c;
  --brand-experiment: #5865f2;
}
</style>

<script>
import { marked, formatText } from "@/utilities";

export default {
  props: {
    avatar: {
      type: String,
      default: "/assets/logo.svg",
    },
    author: {
      type: String,
      default: "Welcomer",
    },
    authorColour: {
      type: Number,
    },
    isBot: {
      type: Boolean,
    },
    content: {
      type: String,
    },
    embeds: {
      type: Object,
    },
    showTimestamp: {
      type: Boolean,
    },
    enableURLs: {
      type: Boolean,
    },
    isDark: {
      type: Boolean,
    },
    respectDarkMode: {
      type: Boolean,
      default: true,
    },
  },
  methods: {
    formatText(text) {
      return formatText(text);
    },

    marked(text, embed) {
        return marked(text, embed);
    },

    rgbIntToRGB(rgbInt, defaultValue) {
      return (
        "#" +
        (rgbInt == undefined ? defaultValue : rgbInt)
          .toString(16)
          .slice(-6)
          .padStart(6, "0")
      );
    },
  },

  setup(props) {
    let now = new Date();

    props.embeds?.forEach((e) => {
      let odd = false;
      e.fields?.forEach((f) => {
        if (f.inline) {
          f.odd = odd;
          odd = !odd;
        } else {
          odd = false;
        }
      });
    });

    return {
      now,
      timestamp: `Today at ${String(now.getHours()).padStart(2, "0") +
        ":" +
        String(now.getMinutes()).padStart(2, "0")
        }`,
    };
  },
};
</script>
