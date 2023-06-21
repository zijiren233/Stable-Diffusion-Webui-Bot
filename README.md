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
  -d    development mode
  -dsn string
        database, postgres|sqlite (default "./stable-diffusion-webui-bot.db")
  -host string
        webhook and api host
  -imgMax int
        image maximum resolution (default 1638400)
  -isp string
        Image Save Path (default "/mnt/stable-diffusion-webui-bot")
  -mf int
        free user max free time
  -owner int
        owner telegram id (default 2143676086)
  -p int
        port (default 8082)
  -t string
        telegram bot token
  -web string
        website site host (default "pyhdxy.top")
  -webhook
        enable telegram bot webhook
```

---

# How to use

- [Download the binary executable file corresponding to the operating system and cpu architecture](https://github.com/zijiren233/Stable-Diffusion-Webui-Bot/releases)
- Copy `config.example.yaml` to `config.yaml` and then configure
- Add running parameters, such as `-t <tg-bot-token>`