package i18n

func init() {
	register(text{language: pt_br, Code: "pt_br", Name: "português"})
}

var pt_br = map[string]string{
	"help":                  "Basta enviar a tag de palavra-chave na caixa de mensagem (apenas em inglês)\nformato de tag: loli,white tights,Uniform\nVocê também pode adicionar uma tag [ao mesmo tempo] como enviar a imagem (a tag está escrita na legenda ao enviar a imagem)\n\nNão use {}, mas use (), () aumenta o peso em 1.1 vezes, e [] reduz o peso em 1.1 vezes.\nUso: (masterpiece:1.1), ((best quality)), algumas tags...\n\nOpções de imagem:\nFT: afinação (quanto maior o número, maior o grau de mudança)\nSPR: super resolução (o número representa a ampliação da imagem)",
	"history":               "Vá para o site para fazer login para ver fotos históricas",
	"setLangSuccess":        "Definir idioma para português",
	"cancel":                "Cancelar",
	"size":                  "Tamanho da imagem",
	"number":                "Número da imagem",
	"mode":                  "modo",
	"unwanted":              "Conteúdo indesejado",
	"confirm":               "Confirme",
	"taskExist":             "Já existe uma tarefa, aguarde a conclusão da tarefa atual",
	"generating":            "Gerando (se esta mensagem não desaparecer, significa que ainda está sendo gerada)\nAguarde pacientemente, se a geração falhar, tente novamente",
	"joinGroup":             "O uso atual é limitado, junte-se ao grupo para aumentar o limite",
	"customUC":              "Por favor, envie o conteúdo que você não quer que apareça (tag reversa):",
	"nsfw":                  "Não é adequado para crianças (Nsfw)",
	"lowQuality":            "Baixa qualidade",
	"badAnatomy":            "Anatomia ruim",
	"none":                  "Nenhum (limpar seleção)",
	"custom":                "Personalizado",
	"strength":              "Força",
	"strengthInfo":          "Controla o quanto a imagem carregada será alterada. Intensidade inferior gerará uma imagem mais próxima da original",
	"serErr":                "O servidor encontrou um erro, tente mais tarde. Ou junte-se ao grupo para discussões e ajuda.",
	"prohibit":              "Você atingiu o limite de uso gratuito para hoje, para continuar usando o bot sem limites por um mês, forneça patrocínio no valor de 3$ dólares ou mais.\nO limite diário será redefinido em {{.time}}\nVocê também pode obter uso extra convidando novos usuários",
	"freeTimes":             "Tempo restante de hoje (entrar no grupo pode aumentar o número de usos gratuitos): ",
	"clickMe":               "Clique para pular",
	"translation":           "Tradução",
	"translate":             "Tag traduz automaticamente para o inglês",
	"reDraw":                "Regeneração",
	"sendTag":               "Por favor, envie a tag:",
	"sendPhoto":             "Por favor, envie a imagem original (não compactada):",
	"parsePhotoErr":         "A análise falhou, envie a imagem mais original (não compactada)...",
	"privateChat":           "Por favor, robô de bate-papo privado",
	"tokenErr":              "Token inválido",
	"model":                 "modelo",
	"scale":                 "aptidão",
	"scaleInfo":             "com que intensidade a imagem deve estar em conformidade com as tags - valores mais baixos produzem resultados mais criativos",
	"steps":                 "degraus",
	"stepsInfo":             "Número de Iterações - Valores mais altos resultam em tempos de compilação mais longos e resultados potencialmente mais detalhados e limpos (e resultados potencialmente piores)",
	"sendImg":               "por favor envie fotos (a resolução total W*H não pode exceder 4194304):",
	"bigImg":                "A resolução da imagem é muito grande (a resolução total W*H não pode exceder 4194304)",
	"magnification":         "Escolha uma ampliação",
	"edit":                  "Editar",
	"modelInfo":             "Diferentes modelos terão grandes diferenças: haverá muitas diferenças no estilo de pintura, personagens, cenários, dimensões, etc..",
	"modeInfo":              "Modos diferentes terão pequenas diferenças na velocidade e nos resultados, sem afetar o estilo e o conteúdo da pintura principal",
	"ucInfo":                "Conteúdo indesejado, geralmente um antônimo de tag",
	"wait":                  "Aguarde um pouco após o envio",
	"dontDelMsg":            "Por favor, não exclua esta mensagem",
	"editTag":               "Editar Tag",
	"Happend":               "inserção de cabeça",
	"Eappend":               "inserção de cauda",
	"setImg":                "definir imagem",
	"setImgInfo":            "Desenhe com base na imagem carregada",
	"clearImg":              "imagem clara",
	"mustShare":             "Apenas usuários inscritos têm permissão para modificar esta configuração, vá para comprar uma assinatura",
	"enable":                "Habilitar",
	"disable":               "Desativado",
	"shareInfo":             "Esta opção determina se a imagem resultante é compartilhada no site, os descadastrados sempre a compartilharão",
	"resetSeed":             "redefinir semente",
	"reset":                 "resetar",
	"extraModel":            "modelo extra",
	"switch":                "trocar",
	"extraModelInfo":        "O modelo extra é apenas para habilitar o modelo relacionado, será inválido se a tag relacionada não for adicionada. Por exemplo, se o modelo Nahida for carregado, você deverá adicionar nahida à tag para entrar em vigor",
	"noSubscribe":           "Atualmente não há assinaturas ativas ou a assinatura expirou",
	"setControl":            "Imagem de controle",
	"editControl":           "Editar gráfico de controle",
	"delControl":            "excluir gráfico de controle",
	"controlPreprocess":     "Pré-processador",
	"controlProcess":        "processador",
	"back":                  "<<< Voltar",
	"setDft":                "Conjunto padrão",
	"onlySubscribe":         "Somente usuários inscritos têm direito de usar, por favor, vá para comprar uma assinatura",
	"canny":                 "detecção de borda",
	"depth":                 "estimativa de mapa de profundidade - MiDaS",
	"depth_leres":           "estimativa de mapa de profundidade - LeReS",
	"hed":                   "detecção de borda suave - HED",
	"hed_safe":              "detecção de borda suave conservadora - HED segura",
	"mediapipe_face":        "detecção de borda facial",
	"mlsd":                  "detecção de linha reta - M-LSD",
	"normal_map":            "extração de mapa normal - Midas",
	"openpose":              "pose - OpenPose",
	"openpose_hand":         "pose | mão - OpenPose",
	"openpose_face":         "pose | rosto - OpenPose",
	"openpose_faceonly":     "somente rosto - OpenPose",
	"openpose_full":         "pose | mão | face - OpenPose",
	"clip_vision":           "processamento de transferência de estilo - adaptativo",
	"color":                 "processamento de pixel de cor - adaptativo",
	"pidinet":               "detecção de borda suave - PiDiNet",
	"pidinet_safe":          "detecção de borda suave conservadora - PiDiNet seguro",
	"pidinet_sketch":        "processamento de borda desenhada à mão - adaptativo",
	"pidinet_scribble":      "rabisco - desenho à mão",
	"scribble_xdog":         "rabisco - reforço de borda",
	"scribble_hed":          "rabisco - composição",
	"threshold":             "limiar",
	"depth_zoe":             "estimativa de mapa de profundidade - ZoE",
	"normal_bae":            "extração de mapa normal - Bae",
	"oneformer_coco":        "segmentação semântica - OneFormer-COCO",
	"oneformer_ade20k":      "segmentação semântica - OneFormer-ADE20K",
	"lineart":               "extração de arte-line",
	"lineart_coarse":        "extração de arte-line grosseira",
	"lineart_anime":         "extração de arte-line de anime",
	"lineart_standard":      "extração padrão de arte-line - inverso",
	"shuffle":               "embaralhamento aleatório",
	"tile_gaussian":         "amostragem por blocos",
	"invert":                "inversão",
	"lineart_anime_denoise": "extração de arte-line de anime - remoção de ruído",
	"reference_only":        "apenas referência da imagem de entrada",
	"inpaint":               "reconstrução - algoritmo de fusão global",
	"invite":                "Quando um novo usuário (que nunca usou ou usou por no máximo 15 minutos) clicar no seu link de convite\nvocê receberá 5 chances extras de uso (que não serão redefinidas e podem ser acumuladas)\nO novo usuário convidado receberá 10 chances extras de uso!",
	"inviteSuccess":         "Você convidou com sucesso o usuário: {{user}}\nGanhou 5 chances extras de uso\ntotal de chances restantes: {{freeAmount}}",
	"wasInvited":            "Você foi convidado pelo usuário: {{user}}, ganhou 10 chances extras de uso\ntotal de chances restantes: {{freeAmount}}",
	"freeMaxNum":            "Usuários gratuitos podem gerar até 3 fotos de cada vez",
}
