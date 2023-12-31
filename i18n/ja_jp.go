package i18n

func init() {
	register(text{language: ja_jp, Code: "ja_jp", Name: "日本"})
}

var ja_jp = map[string]string{
	"help":                  "メッセージボックスにキーワードタグを送信するだけです（英語のみ）\nタグ形式:loli,white tights,Uniform\n画像送信時にタグ付けも【同時に】できます（タグは画像送信時の説明文に書いてあります）\n\n{} は使用せず、() を使用します。() は重みを 1.1 倍にし、[] は重みを 1.1 倍に減らします。\n使用法: (masterpiece:1.1), ((best quality)), some tags...\n\n画像オプション:\nFT: 微調整 (数値が大きいほど変化の度合いが大きくなります)\nSPR: 超解像 (数値は画像の倍率を表します)",
	"history":               "ウェブサイトにアクセスしてログインし、過去の写真を表示します",
	"setLangSuccess":        "言語を日本語に設定",
	"cancel":                "キャンセル",
	"size":                  "絵の大きさ",
	"number":                "写真の数",
	"mode":                  "サンプリングモード",
	"unwanted":              "不要",
	"confirm":               "確認",
	"taskExist":             "既存のタスクが存在します。現在のタスクが終了するまでお待ちください",
	"generating":            "生成中 (このメッセージは消えません。まだ生成中であることを意味します)\nビルドが失敗した場合は、しばらくお待ちください。",
	"joinGroup":             "現在の使用は制限されている、制限を解除するためにグループに参加する",
	"customUC":              "見せたくない内容を送信してください（逆タグ）:",
	"nsfw":                  "お子様には適していません(Nsfw)",
	"lowQuality":            "低品質",
	"badAnatomy":            "間違った構造",
	"none":                  "なし（選択をクリア）",
	"custom":                "カスタマイズ",
	"strength":              "強さ",
	"strengthInfo":          "アップロードされた画像がどの程度変更されるかを制御します。強度を低くすると、よりオリジナルに近い画像が生成されます",
	"serErr":                "サーバーでエラーが発生しました。後で試してください。 または、ディスカッションやヘルプのためにグループに参加してください。",
	"prohibit":              "今日の無料使用制限に達しました。ボットを 1 か月間制限なく使用し続けるには、3 ドル以上のスポンサーシップを提供してください。\n1 日あたりの制限は {{.time}} 後にリセットされます\n新しいユーザーを招待することで追加の使用量を獲得することもできます新しいユーザーを招待することで追加の使用量を獲得することもできます",
	"freeTimes":             "今日の残り時間（グループに参加すると無料利用回数が増える可能性があります）: ",
	"clickMe":               "クリックしてジャンプ",
	"translation":           "翻訳",
	"translate":             "タグ 英語への自動翻訳",
	"reDraw":                "再生",
	"sendTag":               "タグを送信してください:",
	"sendPhoto":             "オリジナル画像（圧縮していないもの）をお送りください。:",
	"parsePhotoErr":         "パースに失敗しました。最もオリジナルの画像（圧縮されていないもの）を送ってください。...",
	"privateChat":           "ロボットと個人的にチャットしてください",
	"tokenErr":              "Token 無効",
	"model":                 "モデル",
	"reset":                 "リセット",
	"scale":                 "適性",
	"scaleInfo":             "画像がタグにどの程度準拠する必要があるか - 値が低いほど、よりクリエイティブな結果になります",
	"steps":                 "歩数",
	"stepsInfo":             "反復回数 - 値が大きいほどビルド時間が長くなり、より詳細でクリーンな結果になる可能性があります (さらに悪い結果になる可能性もあります)。",
	"sendImg":               "写真を送ってください (合計解像度 W*H は 4194304 を超えることはできません):",
	"bigImg":                "画像の解像度が大きすぎます (合計解像度 W*H は 4194304 を超えることはできません)",
	"magnification":         "倍率を選ぶ",
	"edit":                  "編集",
	"modelInfo":             "異なるモデルには大きな違いがあります。塗装スタイル、キャラクター、風景、寸法などに多くの違いがあります。",
	"modeInfo":              "モードが異なると、メインのペイント スタイルやコンテンツに影響を与えることなく、速度と結果がわずかに異なります。",
	"ucInfo":                "望ましくないコンテンツ、通常はタグの反意語",
	"wait":                  "送信後、しばらくお待ちください",
	"dontDelMsg":            "このメッセージを削除しないでください",
	"editTag":               "タグを編集",
	"Happend":               "頭の挿入",
	"Eappend":               "尻尾挿入",
	"setImg":                "セット画像",
	"setImgInfo":            "アップロードした画像をもとに描画",
	"clearImg":              "鮮明な画像",
	"mustShare":             "購読ユーザーのみがこの設定を変更する権限を持っています。購読を購入してください",
	"enable":                "有効",
	"disable":               "無効",
	"shareInfo":             "このオプションは、結果の画像が Web サイトで共有されるかどうかを決定します。",
	"resetSeed":             "シードをリセット",
	"extraModel":            "余分なモデル",
	"switch":                "スイッチ",
	"extraModelInfo":        "エクストラモデルは関連モデルを有効にするだけで、関連タグが追加されていないと無効になります。 たとえば、Nahida モデルがロードされている場合、タグに nahida を追加して有効にする必要があります。",
	"noSubscribe":           "現在アクティブなサブスクリプションがないか、サブスクリプションの有効期限が切れています",
	"setControl":            "コントロール画像",
	"editControl":           "管理図の編集",
	"delControl":            "管理図を削除",
	"controlPreprocess":     "プリプロセッサ",
	"controlProcess":        "プロセッサー",
	"back":                  "<<< 戻る",
	"setDft":                "デフォルトに設定",
	"onlySubscribe":         "サブスクリプションを購入しているユーザーのみが使用できます。サブスクリプションを購入してください。",
	"canny":                 "エッジ検出",
	"depth":                 "深度推定-MiDaS",
	"depth_leres":           "深度推定-LeReS",
	"hed":                   "ソフトエッジ検出-HED",
	"hed_safe":              "セーフソフトエッジ検出-HED",
	"mediapipe_face":        "顔のエッジ検出",
	"mlsd":                  "直線線分検出-M-LSD",
	"normal_map":            "法線マップ抽出-Midas",
	"openpose":              "ポーズ推定-OpenPose",
	"openpose_hand":         "ポーズ推定|手-OpenPose",
	"openpose_face":         "ポーズ推定|顔-OpenPose",
	"openpose_faceonly":     "顔のみ-OpenPose",
	"openpose_full":         "ポーズ推定|手|顔-OpenPose",
	"clip_vision":           "スタイル変換処理-自動適応",
	"color":                 "色彩ドット処理-自動適応",
	"pidinet":               "セーフソフトエッジ検出-PiDiNet",
	"pidinet_safe":          "セーフソフトエッジ検出-PiDiNet",
	"pidinet_sketch":        "手描きエッジ処理-自動適応",
	"pidinet_scribble":      "落書き-手描き",
	"scribble_xdog":         "落書き-強化エッジ",
	"scribble_hed":          "落書き-合成",
	"threshold":             "閾値",
	"depth_zoe":             "深度推定-ZoE",
	"normal_bae":            "法線マップ抽出-Bae",
	"oneformer_coco":        "セマンティックセグメンテーション-OneFormer-COCO",
	"oneformer_ade20k":      "セマンティックセグメンテーション-OneFormer-ADE20K",
	"lineart":               "線画抽出",
	"lineart_coarse":        "荒い線画抽出",
	"lineart_anime":         "アニメ線画抽出",
	"lineart_standard":      "標準線画抽出-反転",
	"shuffle":               "ランダムシャッフル",
	"tile_gaussian":         "ブロックリサンプリング",
	"invert":                "色反転",
	"lineart_anime_denoise": "アニメ線画抽出-ノイズ除去",
	"reference_only":        "参考入力のみ",
	"inpaint":               "再描画-グローバル融合アルゴリズム",
	"invite":                "新規ユーザー（これまでに使用したことがない、または使用時間が15分以内のユーザー）があなたの招待リンクをクリックすると\nあなたは追加で5回の使用機会を得ることができます（リセットされず、蓄積可能）\n招待された新規ユーザーは、追加で10回の使用回数を得ることができます！",
	"inviteSuccess":         "ユーザー{{user}}を招待に成功しました\n追加で5回の使用回数を獲得し\n残りの使用回数：{{freeAmount}}",
	"wasInvited":            "ユーザー{{user}}に招待されました、追加で10回の使用回数を獲得し\n残りの使用回数：{{freeAmount}}",
	"freeMaxNum":            "無料ユーザーは毎回最大 3 枚の写真を生成できます",
}
