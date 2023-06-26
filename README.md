# Stable Diffusion Webui Bot With Telegram

- [Our Chat](https://t.me/SDBot_Chat)
- This is an open source project, no charges are allowed!
- Support for multiple SD API backends
- The Bot interface has perfect operation functions
- Multilingual support
- Owner use `/gettoken 30` to get 30days token
- Recommended Stable Diffusion Webui Start Command Args `export COMMANDLINE_ARGS="--api --no-hashing --skip-torch-cuda-test --skip-version-check --disable-nan-check --no-download-sd-model --no-half-controlnet --upcast-sampling --no-half-vae --opt-sdp-attention --disable-safe-unpickle --lowram --opt-split-attention --opt-channelslast --deepdanbooru"`

---

# Necessary SD Webui extensions

- [Control Net](https://github.com/Mikubill/sd-webui-controlnet)
  - You can run this cmd to install controlnet or update `if [ ! -d "./stable-diffusion-webui/extensions/controlnet" ]; then git clone --depth 1 "https://github.com/Mikubill/sd-webui-controlnet" "./stable-diffusion-webui/extensions/controlnet"; else git -C "./stable-diffusion-webui/extensions/controlnet" pull; fi`

---

# Flags

```
  -api-host string
        set api url (default "127.0.0.1:8082")
  -api-scheme string
        api scheme: http | https (default "http")
  -dev
        development mode
  -dsn string
        database, postgres|sqlite (default "./stable-diffusion-webui-bot.db")
  -i18n-extra-path string
        i18n extra translated (default "./i18n-extra")
  -img-cache-num int
        image cache to mem max num (default 100)
  -img-max int
        image maximum resolution (default 1638400)
  -img-save-path string
        Image Save Path (default "./local-cache")
  -invite
        Enable Invite
  -listen string
        listening address: 127.0.0.1 | 0.0.0.0 (default "127.0.0.1")
  -max-free int
        free user max free time
  -max-num int
        paid user max images (default 6)
  -owner int
        owner telegram id (default 2143676086)
  -port int
        port (default 8082)
  -tg-token string
        telegram bot token
  -webhook-host string
        enable telegram bot webhook: webhook.doamin.com
```

---

# How to use

- [Download the binary executable file corresponding to the operating system and cpu architecture](https://github.com/zijiren233/Stable-Diffusion-Webui-Bot/releases)
- Copy `config.example.yaml` to `config.yaml` and then configure
- Add running parameters, such as `-t <tg-bot-token>`