package majsoul

const (
	ActionDiscard = 1  // 出牌
	ActionChi     = 2  // 吃
	ActionPon     = 3  // 碰
	ActionAnKAN   = 4  // 暗槓
	ActionMinKan  = 5  // 明槓
	ActionKaKan   = 6  // 加槓
	ActionRiichi  = 7  // 立直
	ActionTsumo   = 8  // 自摸
	ActionRon     = 9  // 栄和
	ActionKuku    = 10 // 九九流局
	ActionKita    = 11 // 北
	ActionPass    = 12 // 見逃

	NotifyChi   = 0 // 吃
	NotifyPon   = 1 // 碰
	NotifyKan   = 2 // 杠
	NotifyAnKan = 3 // 暗杠
	NotifyKaKan = 4 // 加杠

	EBakaze = 0 // 东风
	SBakaze = 1 // 南风
	WBakaze = 2 // 西风
	NBakaze = 3 // 北风

	Toncha = 0 // 東家
	Nancha = 1 // 南家
	ShaCha = 2 // 西家
	Peicha = 3 // 北家

	Kyoku1 = 0 // 第1局
	Kyoku2 = 1 // 第2局
	Kyoku3 = 2 // 第3局
	Kyoku4 = 3 // 第4局
)

var Tiles = []string{"1m", "2m", "3m", "4m", "5m", "0m", "6m", "7m", "8m", "9m", "1p", "2p", "3p", "4p", "5p", "0p", "6p", "7p", "8p", "9p", "1s", "2s", "3s", "4s", "5s", "0s", "6s", "7s", "8s", "9s", "1z", "2z", "3z", "4z", "5z", "6z", "7z"}
