package main

import (
	"fmt"
	"log"
	"math/rand"
	"openpenpal-backend/internal/config"
	"openpenpal-backend/internal/models"
	"time"

	"gorm.io/gorm"
)

// 信件内容模板
var letterTemplates = []struct {
	Title   string
	Content string
	Style   string
	Tags    []string
}{
	{
		Title: "写给三年后的自己",
		Content: `亲爱的未来的我，

当你读到这封信的时候，希望你已经成为了更好的自己。现在的我正站在人生的十字路口，对未来充满期待却又略带迷茫。

我想对三年后的你说：
1. 希望你还记得当初的梦想，并且正在为之努力
2. 希望你学会了与孤独相处，也找到了志同道合的朋友
3. 希望你变得更加勇敢，不再害怕失败
4. 希望你保持着对生活的热爱，对世界的好奇

无论你现在在哪里，做着什么，请记住：你已经很棒了。

爱你的，
过去的自己`,
		Style: "future",
		Tags:  []string{"成长", "梦想", "大学生活", "未来"},
	},
	{
		Title: "致正在迷茫的你",
		Content: `如果你正在经历人生的低谷，请记住这只是暂时的。

每个人都会有迷茫的时候，这很正常。迷茫意味着你在思考，在寻找属于自己的方向。

我想告诉你：
- 不要急着找到所有答案，人生本就是一个不断探索的过程
- 允许自己偶尔脆弱，这不是软弱，而是真实
- 相信时间的力量，很多事情会慢慢变好
- 记得照顾好自己，身心健康最重要

黑夜再长，黎明终会到来。你比自己想象的要坚强。

一个陌生的朋友`,
		Style: "warm",
		Tags:  []string{"鼓励", "治愈", "心理健康", "温暖"},
	},
	{
		Title: "一个关于友谊的故事",
		Content: `我想和你分享一个关于友谊的故事。

大学第一天，我遇到了室友小李。起初我们并不熟悉，甚至因为生活习惯的差异有过小摩擦。但是一次深夜，当我因为挂科而崩溃大哭时，是他默默地陪在我身边，递给我纸巾，什么也没说。

从那以后，我们成了最好的朋友。一起刷夜复习，一起逃课看电影，一起在食堂吐槽难吃的饭菜，一起规划着毕业后的生活。

友谊就是这样，它不需要轰轰烈烈，只需要在彼此需要的时候，有一个人在身边。

谢谢你，我的朋友。`,
		Style: "story",
		Tags:  []string{"友谊", "青春", "回忆", "大学"},
	},
	{
		Title: "漂流到远方的思念",
		Content: `这封信将随风漂流到某个角落，希望能遇到同样思念远方的你。

离开家乡来到这座陌生的城市已经两年了。每当夜深人静的时候，总会想起家乡的味道——妈妈做的红烧肉，爸爸泡的茶，还有奶奶院子里的桂花香。

思念是一种很奇妙的东西，它让远方变得更远，却也让心与心的距离变得更近。

如果你也在思念着什么人或什么地方，请记住：思念是爱的另一种表达方式。

愿这封信能带给你一丝温暖。`,
		Style: "drift",
		Tags:  []string{"思念", "漂流", "家乡", "温情"},
	},
	{
		Title: "大学四年的感悟",
		Content: `即将毕业，回想这四年的大学时光，有太多话想说。

大一的青涩，大二的迷茫，大三的忙碌，大四的不舍。四年时间说长不长，说短不短，却足以改变一个人。

我学会了：
- 独立思考比盲从更重要
- 真正的朋友不在多，而在精
- 失败是成功的必经之路
- 要勇于走出舒适区
- 珍惜当下的每一刻

致所有正在经历大学生活的你们：好好享受这段时光吧，它真的一去不复返。

一个即将毕业的学长`,
		Style: "story",
		Tags:  []string{"毕业", "感悟", "大学", "成长"},
	},
	{
		Title: "写给十年后的世界",
		Content: `2034年的世界会是什么样子？

我猜想：
- AI已经成为生活的一部分，但人类的创造力依然无可替代
- 环保不再是口号，而是每个人的生活方式
- 远程工作成为常态，地理距离不再是障碍
- 医疗技术的进步让更多疾病得到治愈
- 人们学会了在快节奏中寻找慢生活

但无论科技如何发展，我相信人与人之间的温情不会改变，文字的力量依然存在。

十年后的你，还在写信吗？`,
		Style: "future",
		Tags:  []string{"未来", "科技", "想象", "展望"},
	},
	{
		Title: "深夜食堂的温暖",
		Content: `凌晨两点，学校附近的小食堂依然灯火通明。

老板是个胖胖的中年人，总是笑眯眯的。这里是无数个深夜的避风港——有刚下晚自习的学生，有加班归来的上班族，有失恋买醉的年轻人。

一碗热腾腾的面条，一句"辛苦了"，就能治愈一天的疲惫。

在这个小小的食堂里，我看到了生活最真实的样子：有苦有甜，有泪有笑，但总有一盏灯为你点亮。

谢谢你，深夜食堂。`,
		Style: "warm",
		Tags:  []string{"温暖", "深夜", "美食", "人情味"},
	},
	{
		Title: "第一次一个人旅行",
		Content: `背上背包，买了一张单程票，开始了我的第一次独自旅行。

没有详细的攻略，没有同伴，只有一颗想要看看世界的心。在陌生的城市迷路，用蹩脚的英语问路，在青旅遇到来自世界各地的朋友，在山顶看日出...

一个人旅行教会了我：
- 独处也可以很精彩
- 勇气比完美的计划更重要
- 世界很大，但人心相通
- 每一次出发都是成长

如果你也想尝试，就勇敢地出发吧！`,
		Style: "story",
		Tags:  []string{"旅行", "成长", "勇气", "独立"},
	},
	{
		Title: "雨夜的诗",
		Content: `窗外雨声淅沥，
思绪如这雨丝般绵长。

想起李商隐的"君问归期未有期，巴山夜雨涨秋池"，
想起戴望舒的"撑着油纸伞，独自彷徨在悠长、悠长又寂寥的雨巷"。

雨夜适合思念，
适合写诗，
适合做梦。

在这个雨夜，
我把思念寄给远方，
把诗意留在心上。

晚安，听雨的人。`,
		Style: "drift",
		Tags:  []string{"诗意", "雨夜", "思念", "文艺"},
	},
	{
		Title: "那些年我们一起追过的梦",
		Content: `还记得高中时，我们趴在栏杆上聊着各自的梦想吗？

你说要当医生，救死扶伤；
他说要做老师，桃李满天下；
她说要环游世界，看遍风景；
我说要写一本书，记录青春。

现在，我们都在各自的路上努力着。虽然现实和梦想有差距，但至少我们都在前进。

致所有还在追梦路上的朋友们：
梦想也许会迟到，但永远不会缺席。

加油！`,
		Style: "warm",
		Tags:  []string{"梦想", "青春", "友谊", "励志"},
	},
}

// 用户列表（使用系统中已存在的测试用户）
var usernames = []string{
	"alice",
	"bob",
	"charlie",
	"david",
	"eve",
	"frank",
	"grace",
	"henry",
	"iris",
	"jack",
}

func main() {
	// 加载配置
	cfg := config.LoadConfig()
	db := config.InitDB(cfg)

	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	// 清理现有的公开信件（可选）
	var deleteExisting string
	fmt.Print("是否删除现有的公开信件？(y/N): ")
	fmt.Scanln(&deleteExisting)
	if deleteExisting == "y" || deleteExisting == "Y" {
		db.Where("visibility = ?", "public").Delete(&models.Letter{})
		fmt.Println("已删除现有的公开信件")
	}

	// 获取用户ID映射
	userIDMap := make(map[string]string)
	for _, username := range usernames {
		var user models.User
		if err := db.Where("username = ?", username).First(&user).Error; err == nil {
			userIDMap[username] = user.ID
		} else {
			fmt.Printf("警告：用户 %s 不存在，跳过该用户\n", username)
		}
	}

	if len(userIDMap) == 0 {
		log.Fatal("没有找到任何有效用户，请先运行用户种子脚本")
	}

	// 创建信件
	successCount := 0
	for i := 0; i < 3; i++ { // 每个模板创建3封信
		for _, template := range letterTemplates {
			// 随机选择一个用户
			var userID string
			var username string
			for u, id := range userIDMap {
				username = u
				userID = id
				if rand.Float32() < 0.3 { // 30%概率选中
					break
				}
			}

			// 生成信件数据
			letter := models.Letter{
				UserID:       userID,
				Title:        fmt.Sprintf("%s (%d)", template.Title, i+1),
				Content:      template.Content,
				Style:        template.Style,
				PaperType:    getRandomPaperType(),
				FontStyle:    getRandomFontStyle(),
				Status:       models.LetterStatusPublished,
				Visibility:   "public",
				AllowComment: true,
				IsAnonymous:  rand.Float32() < 0.3, // 30%概率匿名
				CreatedAt:    getRandomPastDate(),
				UpdatedAt:    time.Now(),
			}

			// 生成信件代码
			letter.GenerateCode()

			// 添加随机的互动数据
			letter.ViewCount = rand.Intn(2000) + 100
			letter.LikeCount = rand.Intn(300) + 10
			letter.CommentCount = rand.Intn(50) + 0
			letter.ShareCount = rand.Intn(20) + 0

			// 创建信件
			if err := db.Create(&letter).Error; err != nil {
				fmt.Printf("创建信件失败 (%s): %v\n", template.Title, err)
				continue
			}

			successCount++
			fmt.Printf("✓ 创建信件: %s (作者: %s, 代码: %s)\n", letter.Title, username, letter.Code)

			// 为热门信件添加更多互动数据
			if rand.Float32() < 0.3 { // 30%概率成为热门
				letter.ViewCount += rand.Intn(3000)
				letter.LikeCount += rand.Intn(500)
				letter.CommentCount += rand.Intn(100)
				letter.ShareCount += rand.Intn(50)
				db.Save(&letter)
			}

			// 随机添加一些点赞记录
			if rand.Float32() < 0.5 {
				for _, likerUsername := range usernames {
					if rand.Float32() < 0.2 && likerUsername != username {
						if likerID, ok := userIDMap[likerUsername]; ok {
							like := models.LetterLike{
								LetterID:  letter.ID,
								UserID:    likerID,
								CreatedAt: getRandomPastDate(),
							}
							db.Create(&like)
						}
					}
				}
			}
		}
	}

	fmt.Printf("\n✅ 成功创建 %d 封公开信件！\n", successCount)
	fmt.Println("\n现在可以访问 http://localhost:3000/plaza 查看写作广场")
}

// 辅助函数
func getRandomPaperType() string {
	types := []string{"classic", "modern", "vintage", "elegant", "casual"}
	return types[rand.Intn(len(types))]
}

func getRandomFontStyle() string {
	styles := []string{"handwritten", "print", "cursive", "regular"}
	return styles[rand.Intn(len(styles))]
}

func getRandomPastDate() time.Time {
	// 生成过去30天内的随机日期
	days := rand.Intn(30) + 1
	hours := rand.Intn(24)
	return time.Now().Add(-time.Duration(days*24+hours) * time.Hour)
}