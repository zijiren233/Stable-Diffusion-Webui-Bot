# Stable Diffusion Webui Bot With Telegram

- [Our Chat](https://t.me/SDBot_Chat)
- This is an open source project, no charges are allowed!
- Owner use `/gettoken 30` to get 30days token
- Recommended Stable Diffusion Webui Start Command Args `export COMMANDLINE_ARGS="--api --no-hashing --skip-torch-cuda-test --skip-version-check --disable-nan-check --no-download-sd-model --no-half-controlnet --upcast-sampling --no-half-vae --opt-sdp-attention --disable-safe-unpickle --lowram --opt-split-attention --opt-channelslast --deepdanbooru"`
- The necessary extensions
  - `https://github.com/Mikubill/sd-webui-controlnet`

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