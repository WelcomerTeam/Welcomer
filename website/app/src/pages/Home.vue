<template>
  <div class="relative min-h-screen bg-secondary">
    <Header />
    <HoistHeading />
    <main>
      <div class="mx-auto my-4 lg:relative max-w-7xl">
        <div class="hero">
          <div class="px-6 my-auto lg:w-1/2 sm:px-8 xl:pr-16">
            <h1 class="hero-title">The go-to bot for your discord server</h1>
            <p class="hero-subtitle">
              Providing tools to let you manage your server. Join the more than
              <b>521,473</b> Discord guilds using Welcomer.
            </p>
            <div class="hero-buttons">
              <div class="hero-primary">
                <router-link :to="{ name: 'invite' }">
                  Invite Welcomer
                </router-link>
              </div>
              <div class="hero-secondary">
                <a href="#features">Features</a>
              </div>
            </div>
          </div>
        </div>
        <div class="hero-accompany">
          <div class="flex w-full h-full">
            <DiscordEmbed class="max-w-full m-auto" :isBot="true" :isDark="true" :embeds="[
              {
                description: `Hey **ImRock**, you are the 256th member on **Welcomer Support Guild**`,
                image: {
                  url: '/assets/welcomerImage.svg',
                },
              },
            ]" />
          </div>
        </div>
      </div>

      <div class="bg-white text-secondary">
        <div class="hero-preview">
          <div class="pb-12 prose-lg text-center">
            <h1 class="text-lg font-black leading-8 tracking-tight text-gray-900">
              Elevate Your User Engagement
            </h1>
            <span class="text-lg text-gray-700 max-w-3xl mx-auto section-subtitle">Explore the reasons why Welcomer is
              the go-to choice for empowering Discord guilds. Packed with numerous features and boasting a user-friendly
              dashboard, you have the flexibility to make changes whenever you desire.
            </span>
          </div>

          <div v-for="item in previews" :key="item.name" class="hero-preview-item">
            <div class="my-auto text-center sm:text-left">
              <div>
                <h2 class="mb-4 text-3xl font-black leading-tight text-gray-900">
                  {{ item.name }}
                </h2>
                <span class="text-gray-700">{{ item.description }}
                  <router-link :to="item.href" v-if="item.href"
                    class="text-primary hover:text-primary-dark font-semibold">
                    {{ item.label }}
                  </router-link>
                </span>
              </div>
            </div>
            <div class="my-auto">
              <DiscordEmbed v-if="item.embeds" class="max-w-full m-auto" :isBot="true" :isDark="false" :respectDarkMode="false"
                :embeds="item.embeds" />
              <img v-else :src="item.src" alt="Preview image" class="max-w-full m-auto" />
            </div>
          </div>
        </div>
      </div>

      <div class="bg-primary">
        <div class="hero-features">
          <div class="hero-features-item">
            <div v-for="item in features" :key="item.name"
              class="p-4 mx-0 my-3 text-center rounded-lg sm:text-left sm:mx-2 hover:bg-primary-light">
              <h2 class="text-xl font-medium text-white">{{ item.name }}</h2>
              <p>{{ item.description }}</p>
            </div>
          </div>
        </div>
      </div>

      <div class="bg-secondary">
        <div class="px-4 py-12 mx-auto text-center max-w-prose sm:px-6 md:py-16 lg:px-7 lg:py-20">
          <h2 class="text-2xl font-bold tracking-tight text-gray-900 sm:text-4xl">
            <span class="block text-white">Ready to get started?</span>
            <span class="block text-primary">Invite Welcomer today, it's free.</span>
          </h2>
          <p class="mt-5 text-lg leading-6 text-gray-400">
            Elevate Moderation, Enhance Engagement: Welcomer is your key to a thriving community. Start today!
          </p>
          <router-link :to="{ name: 'dashboard.guilds' }">
            <button type="button" class="w-32 mt-4 cta-button bg-primary hover:bg-primary-dark">
              Dashboard
            </button>
          </router-link>
        </div>
      </div>
    </main>

    <Footer />
  </div>
</template>

<script>
import Header from "@/components/Header.vue";
import Footer from "@/components/Footer.vue";
import DiscordEmbed from "@/components/DiscordEmbed.vue";
import HoistHeading from "@/components/hoist/HoistHeading.vue";

function makeAFunEmbed() {
  let messages = [
    "Welcome to our server! Make sure to checkout our #rules and enjoy your stay!",
    "Welcome aboard! Don't forget to say hi in the #introductions channel. Enjoy your time here!",
    "Glad to have you with us! Check out #announcements for the latest news and have fun!",
    "Welcome to our community! Make sure to explore our channels and join the conversations.",
    "¡Bienvenido a nuestro servidor! No olvides revisar #reglas y disfruta tu estadía.",
    "¡Nos alegra tenerte aquí! Pasa por #presentaciones y cuéntanos sobre ti. ¡Diviértete!",
    "¡Bienvenido a nuestra comunidad! Explora nuestros canales y únete a las conversaciones.",
    "Bienvenue sur notre serveur ! N'oubliez pas de consulter #règles et profitez de votre séjour.",
    "Ravi de vous avoir avec nous ! Présentez-vous dans #présentations et amusez-vous bien !",
    "Bienvenue dans notre communauté ! Explorez nos canaux et rejoignez les discussions.",
    "Willkommen auf unserem Server! Schau dir #regeln an und hab eine gute Zeit!",
    "Schön, dass du da bist! Stell dich in #vorstellungen vor und viel Spaß!",
    "Willkommen in unserer Community! Erkunde unsere Kanäle und mach bei den Gesprächen mit.",
    "サーバーへようこそ！#ルール を確認して、楽しんでくださいね。",
    "ようこそ！#自己紹介 で挨拶して、楽しい時間を過ごしてください！",
    "コミュニティへようこそ！チャンネルを探索して、会話に参加しましょう。",
    "Bem-vindo ao nosso servidor! Não se esqueça de conferir #regras e aproveite sua estadia!",
    "Ficamos felizes em tê-lo conosco! Passe por #apresentações e se divirta!",
    "Bem-vindo à nossa comunidade! Explore nossos canais e junte-se às conversas.",
    "Benvenuto nel nostro server! Non dimenticare di dare un'occhiata a #regole e goditi il tuo soggiorno.",
    "Siamo felici di averti con noi! Presentati in #presentazioni e divertiti!",
    "Benvenuto nella nostra comunità! Esplora i nostri canali e partecipa alle conversazioni.",
    "हमारे सर्वर में आपका स्वागत है! #नियम चैनल को देखना न भूलें और यहाँ का आनंद लें।",
    "हमारे साथ जुड़ने के लिए धन्यवाद! #परिचय में जाकर हमें अपने बारे में बताएं और मज़े करें।",
    "हमारे समुदाय में आपका स्वागत है! हमारे चैनलों को एक्सप्लोर करें और बातचीत में शामिल हों।",
    "ہمارے سرور پر خوش آمدید! #قواعد کو دیکھنا نہ بھولیں اور یہاں کا مزہ لیں۔",
    "ہمارے ساتھ شامل ہونے کا شکریہ! #تعارف میں جائیں اور ہمیں اپنے بارے میں بتائیں اور لطف اندوز ہوں۔",
    "ہماری کمیونٹی میں خوش آمدید! ہمارے چینلز کو دریافت کریں اور بات چیت میں شامل ہوں۔"
  ];

  let thumbnailURLs = [
    "/assets/comfy_white.png",
    "/assets/crown.gif",
    "/assets/explode.gif",
    "/assets/green_gibheart.png",
    "/assets/green_sunglas.png",
    "/assets/red_mlg.png",
    "/assets/red_waving.png",
    "/assets/wave.gif",
    "/assets/white_comfee.png",
  ];

  let imageURLs = [
    "/assets/ejected.png"
  ];

  
  let hslToInt = (h, s, l) => {
    h /= 360;
    s /= 100;
    l /= 100;
  
    let r, g, b;
  
    if (s === 0) {
      r = g = b = l;
    } else {
      const hue2rgb = (p, q, t) => {
        if (t < 0) t += 1;
        if (t > 1) t -= 1;
        if (t < 1 / 6) return p + (q - p) * 6 * t;
        if (t < 1 / 2) return q;
        if (t < 2 / 3) return p + (q - p) * (2 / 3 - t) * 6;
        return p;
      };
  
      const q = l < 0.5 ? l * (1 + s) : l + s - l * s;
      const p = 2 * l - q;
  
      r = hue2rgb(p, q, h + 1 / 3);
      g = hue2rgb(p, q, h);
      b = hue2rgb(p, q, h - 1 / 3);
    }
  
    const toInt = (c) => {
      return Math.round(c * 255);
    };
  
    const intR = toInt(r);
    const intG = toInt(g);
    const intB = toInt(b);
  
    return (intR << 16) + (intG << 8) + intB;
  };

  let embed = {
    description: messages[Math.floor(Math.random() * messages.length)],
    color: hslToInt(Math.floor(Math.random() * 361), 50, 50),
  }

  if (Math.random() < 0.1) {
    embed.image = {
      url: imageURLs[Math.floor(Math.random() * imageURLs.length)],
    };
  } else {
    embed.thumbnail = {
      url: thumbnailURLs[Math.floor(Math.random() * thumbnailURLs.length)],
    };
  }

  return [embed];
}

const previews = [
  {
    name: "Empower your Welcome Messages",
    description: "Tailor your welcome messages with Welcomer, offering personalized customization for text, direct messages, and images, ensuring a unique and engaging experience for every Discord user.",
    embeds: makeAFunEmbed(),
  },
  {
    name: "Reinforce Your Server's Defenses with Borderwall",
    description: "With Borderwall enabled, ensure only authentic users can send messages by presenting a challenge upon joining, guaranteeing a secure and spam-resistant community.",
    src: "/assets/verify.png",
  },
  // {
  //   name: "Increase user engagement",
  //   description:
  //     "Reward users with XP for conversing on your server and let them compete against each other with a public leaderboard.",
  //   src: "/assets/placeholder.png",
  // },
  // {
  //   name: "Automate your server",
  //   description:
  //     "Automation allows you to create tasks that execute on events in a zero-code environment. Want to give roles to anybody who sends “pog”? You can do that without any knowledge of programming.",
  //   src: "/assets/placeholder.png",
  // },
];

const features = [
  // AutoReminders
  {
    name: "AutoRoles",
    description: "Automatically give users roles when they join your server.",
  },
  {
    name: "Borderwall",
    description: "Secure your server from automated accounts with a challenge when joining.",
  },
  {
    name: "FreeRoles",
    description: "Allow users to easily give and remove certain roles from themselves through a command.",
  },
  // InviteRoles
  {
    name: "Leaver",
    description: "Send a message when users leave your server, wishing them farewell.",
  },
  // LevelRoles
  // Lockdown
  // Polls
  // ReactionRoles
  {
    name: "Rules",
    description: "Provide a list of rules for users to see, and send them when a user joins your server.",
  },
  // Starboard
  // StickyRoles
  // Suggestions
  {
    name: "TempChannels",
    description: "Allow users to create temporary voice channels, freeing up space.",
  },
  {
    name: "TimeRoles",
    description: "Reward users for staying in your server for a period of time, with special roles.",
  },
  {
    name: "Welcomer",
    description: "Welcome new users to your servers to your server with fancy images and text or send them a direct message.",
  },
  // XP
];

export default {
  components: {
    Header,
    Footer,
    DiscordEmbed,
    HoistHeading,
  },
  setup() {
    return {
      previews,
      features,
    };
  },
};
</script>
