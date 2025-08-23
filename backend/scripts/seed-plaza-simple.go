package main

import (
	"fmt"
	"log"
	"math/rand"
	"openpenpal-backend/internal/models"
	"time"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"github.com/google/uuid"
)

// 信件内容模板
var letterTemplates = []struct {
	Title   string
	Content string
	Style   string
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
	},
	{
		Title: "一个关于友谊的故事",
		Content: `我想和你分享一个关于友谊的故事。

大学第一天，我遇到了室友小李。起初我们并不熟悉，甚至因为生活习惯的差异有过小摩擦。但是一次深夜，当我因为挂科而崩溃大哭时，是他默默地陪在我身边，递给我纸巾，什么也没说。

从那以后，我们成了最好的朋友。一起刷夜复习，一起逃课看电影，一起在食堂吐槽难吃的饭菜，一起规划着毕业后的生活。

友谊就是这样，它不需要轰轰烈烈，只需要在彼此需要的时候，有一个人在身边。

谢谢你，我的朋友。`,
		Style: "story",
	},
	{
		Title: "漂流到远方的思念",
		Content: `这封信将随风漂流到某个角落，希望能遇到同样思念远方的你。

离开家乡来到这座陌生的城市已经两年了。每当夜深人静的时候，总会想起家乡的味道——妈妈做的红烧肉，爸爸泡的茶，还有奶奶院子里的桂花香。

思念是一种很奇妙的东西，它让远方变得更远，却也让心与心的距离变得更近。

如果你也在思念着什么人或什么地方，请记住：思念是爱的另一种表达方式。

愿这封信能带给你一丝温暖。`,
		Style: "drift",
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
	},
}

func main() {
	// 连接数据库
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// macOS 使用系统用户名
		user := os.Getenv("USER")
		if user == "" {
			user = "postgres"
		}
		dsn = fmt.Sprintf("postgres://%s:@localhost:5432/openpenpal?sslmode=disable", user)
	}
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// 设置随机种子
	rand.Seed(time.Now().UnixNano())

	// 获取测试用户
	var users []models.User
	if err := db.Where("role IN ?", []string{"student", "courier_level1", "courier_level2"}).Limit(10).Find(&users).Error; err != nil {
		log.Fatal("Failed to get users:", err)
	}

	if len(users) == 0 {
		log.Fatal("No users found. Please create test users first.")
	}

	// 清理现有的公开信件（可选）
	var deleteExisting string
	fmt.Print("是否删除现有的公开信件？(y/N): ")
	fmt.Scanln(&deleteExisting)
	if deleteExisting == "y" || deleteExisting == "Y" {
		db.Where("visibility = ?", "public").Delete(&models.Letter{})
		fmt.Println("已删除现有的公开信件")
	}

	// 创建信件
	successCount := 0
	for i := 0; i < 3; i++ { // 每个模板创建3封信
		for _, template := range letterTemplates {
			// 随机选择一个用户
			user := users[rand.Intn(len(users))]

			// 生成信件数据
			letterID := uuid.New().String()
			letter := models.Letter{
				ID:          letterID,
				UserID:      user.ID,
				AuthorID:    user.ID, // 设置 AuthorID 避免外键约束错误
				Title:       fmt.Sprintf("%s (%d)", template.Title, i+1),
				Content:     template.Content,
				Style:       models.LetterStyle(template.Style),
				Status:      models.StatusApproved, // 公开信件应该是已审核状态
				Visibility:  models.VisibilityPublic,
				Type:        models.LetterTypeOriginal,
				AuthorName:  user.Username,
				CreatedAt:   getRandomPastDate(),
				UpdatedAt:   time.Now(),
			}

			// 添加随机的互动数据
			letter.ViewCount = rand.Intn(2000) + 100
			letter.LikeCount = rand.Intn(300) + 10
			letter.ShareCount = rand.Intn(20) + 0

			// 创建信件
			if err := db.Create(&letter).Error; err != nil {
				fmt.Printf("创建信件失败 (%s): %v\n", template.Title, err)
				continue
			}

			successCount++
			
			// 为信件创建一个 LetterCode
			letterCode := models.LetterCode{
				ID:       uuid.New().String(),
				LetterID: letter.ID,
				Code:     generateLetterCode(), // 生成12位代码
				Status:   models.BarcodeStatusBound,
			}
			db.Create(&letterCode)
			
			fmt.Printf("✓ 创建信件: %s (作者: %s, 代码: %s)\n", letter.Title, user.Username, letterCode.Code)

			// 为热门信件添加更多互动数据
			if rand.Float32() < 0.3 { // 30%概率成为热门
				letter.ViewCount += rand.Intn(3000)
				letter.LikeCount += rand.Intn(500)
				letter.ShareCount += rand.Intn(50)
				db.Save(&letter)
			}

			// 随机添加一些点赞记录
			if rand.Float32() < 0.5 && len(users) > 1 {
				likerCount := rand.Intn(5) + 1
				for j := 0; j < likerCount && j < len(users); j++ {
					liker := users[rand.Intn(len(users))]
					if liker.ID != user.ID {
						like := models.LetterLike{
							LetterID:  letter.ID,
							UserID:    liker.ID,
							CreatedAt: getRandomPastDate(),
						}
						db.Create(&like)
					}
				}
			}
		}
	}

	fmt.Printf("\n✅ 成功创建 %d 封公开信件！\n", successCount)
	fmt.Println("\n现在可以访问 http://localhost:3000/plaza 查看写作广场")
}

// 辅助函数
func generateLetterCode() string {
	// 生成12位代码，格式: YYYYMMDDXXXX
	now := time.Now()
	prefix := now.Format("20060102")
	suffix := fmt.Sprintf("%04d", rand.Intn(10000))
	return prefix + suffix
}

func getRandomPastDate() time.Time {
	// 生成过去30天内的随机日期
	days := rand.Intn(30) + 1
	hours := rand.Intn(24)
	return time.Now().Add(-time.Duration(days*24+hours) * time.Hour)
}