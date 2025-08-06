#!/usr/bin/env node

/**
 * Test script to add sample museum data to OpenPenPal database
 * Creates test letters and submits them to the museum
 */

const axios = require('axios');

// Configuration
const API_BASE = 'http://localhost:8080';
const AUTH_TOKEN = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiJ0ZXN0LWFkbWluIiwicm9sZSI6InN1cGVyX2FkbWluIiwiaXNzIjoib3BlbnBlbnBhbCIsImV4cCI6MTc1NDEyMDYzNywiaWF0IjoxNzU0MDM0MjM3LCJqdGkiOiI1NDZkNTYzZTc1ZDJlMzg2NDVkOGQyNmQ1NjBiOTY5ZiJ9.XhP-0k-UhHRZQ3pQk6oN3L8q-tkzkReuTNyFT43qdWU';

// Set NO_PROXY to bypass proxy issues
process.env.NO_PROXY = 'localhost,127.0.0.1';

// Create axios instance with default headers
const api = axios.create({
  baseURL: API_BASE,
  headers: {
    'Authorization': `Bearer ${AUTH_TOKEN}`,
    'Content-Type': 'application/json'
  },
  proxy: false
});

// Sample letter content data
const sampleLetters = [
  {
    title: "春天的记忆",
    content: `亲爱的朋友，

春天来了，校园里的樱花开得正盛。每当微风吹过，粉色的花瓣如雪花般飘落，铺满了林荫道。

我想起了我们一起在樱花树下读书的日子，那些美好的时光永远留在心中。希望你在远方一切安好，期待我们再次相聚的那一天。

愿春天的美好与你同在！

你的朋友`,
    style: "elegant",
    tags: ["春天", "樱花", "友谊", "思念"],
    museumTitle: "樱花树下的友谊",
    museumDescription: "一封充满春天气息的信件，记录了校园里樱花盛开的美景和珍贵的友谊。"
  },
  {
    title: "深夜的思考",
    content: `致未来的自己，

今夜无眠，坐在宿舍的书桌前，窗外是静谧的校园。月光洒在桌面上，陪伴着我的思绪。

大学生活已经过半，我开始思考自己的方向。是继续深造，还是步入社会？每个选择都像是人生的岔路口。

但我相信，无论选择哪条路，只要坚持初心，勇敢前行，终会找到属于自己的光芒。

加油，未来的我！`,
    style: "modern",
    tags: ["深夜", "思考", "未来", "成长"],
    museumTitle: "月光下的人生思考",
    museumDescription: "一位大学生在深夜对未来的思考和自我激励，展现了年轻人的迷茫与坚定。"
  },
  {
    title: "食堂的故事",
    content: `亲爱的学弟学妹们，

作为一个即将毕业的老学长，我想和你们分享一些关于食堂的小秘密。

二食堂的麻辣香锅是最棒的，记得要在11:30之前去，不然就要排很长的队。三食堂二楼的糖醋排骨简直是人间美味，每周三才有哦！

还有，一食堂阿姨人超好，如果你说"阿姨辛苦了"，她会多给你一勺菜的（偷偷告诉你的）。

希望这些小tips能让你们的大学生活更加美好！

一个吃货学长`,
    style: "casual",
    tags: ["美食", "食堂", "校园生活", "传承"],
    museumTitle: "食堂美食攻略",
    museumDescription: "一位即将毕业的学长分享的食堂美食秘籍，充满了对校园生活的热爱。"
  },
  {
    title: "图书馆的邂逅",
    content: `致那个在图书馆遇见的你，

今天下午，在图书馆三楼靠窗的位置，我们的目光不经意地相遇了。你在看《百年孤独》，阳光透过窗户洒在你的侧脸上。

我想上前打招呼，却又不知该说什么。最后只是微微一笑，继续埋头看书。

也许这就是大学里最美好的瞬间——不需要言语，只是静静地存在于同一个空间，各自追寻着知识的光芒。

希望明天还能在那里遇见你。

一个羞涩的读者`,
    style: "vintage",
    tags: ["图书馆", "邂逅", "青春", "浪漫"],
    museumTitle: "图书馆的浪漫邂逅",
    museumDescription: "记录了一次图书馆里的美好邂逅，展现了大学生活中的青涩与浪漫。"
  },
  {
    title: "实验室的不眠夜",
    content: `致我的实验伙伴们，

又是一个通宵的夜晚，实验室的灯光依旧明亮。咖啡已经喝了三杯，数据还在跑着。

虽然很累，但看着你们认真的样子，心中充满了力量。我们一起调试代码，一起分析数据，一起为了那个可能改变世界的想法而努力。

这就是青春该有的样子吧——为了梦想，可以不眠不休。

感谢有你们的陪伴，让这段艰难的科研路不再孤单。

永远的队友`,
    style: "modern",
    tags: ["科研", "团队", "奋斗", "青春"],
    museumTitle: "实验室里的青春奋斗",
    museumDescription: "展现了一群年轻科研人员在实验室通宵奋斗的场景，充满了对梦想的执着。"
  }
];

// Helper function to create a letter
async function createLetter(letterData) {
  try {
    console.log(`Creating letter: "${letterData.title}"...`);
    
    const response = await api.post('/api/v1/letters', {
      title: letterData.title,
      content: letterData.content,
      style: letterData.style
    });
    
    if (response.data.code === 0 || response.data.success) {
      const letter = response.data.data;
      console.log(`✓ Letter created successfully: ${letter.id}`);
      return letter;
    } else {
      console.error(`✗ Failed to create letter: ${response.data.message}`);
      return null;
    }
  } catch (error) {
    console.error(`✗ Error creating letter:`, error.response?.data || error.message);
    return null;
  }
}

// Helper function to generate letter code
async function generateLetterCode(letterId) {
  try {
    console.log(`Generating code for letter: ${letterId}...`);
    
    const response = await api.post(`/api/v1/letters/${letterId}/generate-code`);
    
    if (response.data.code === 0 || response.data.success) {
      console.log(`✓ Letter code generated successfully`);
      return true;
    } else {
      console.error(`✗ Failed to generate letter code: ${response.data.message}`);
      return false;
    }
  } catch (error) {
    console.error(`✗ Error generating letter code:`, error.response?.data || error.message);
    return false;
  }
}

// Helper function to submit letter to museum
async function submitToMuseum(letter, letterData) {
  try {
    console.log(`Submitting letter to museum: "${letterData.museumTitle}"...`);
    
    const response = await api.post('/api/v1/museum/items', {
      sourceType: "letter",
      sourceId: letter.id,
      title: letterData.museumTitle,
      description: letterData.museumDescription,
      tags: letterData.tags,
      submittedBy: "test-admin" // This will be overridden by the handler to use JWT user ID
    });
    
    if (response.data.success) {
      console.log(`✓ Letter submitted to museum successfully`);
      return response.data.data;
    } else {
      console.error(`✗ Failed to submit to museum: ${response.data.message}`);
      return null;
    }
  } catch (error) {
    console.error(`✗ Error submitting to museum:`, error.response?.data || error.message);
    return null;
  }
}

// Helper function to approve museum item (admin only)
async function approveMuseumItem(itemId) {
  try {
    console.log(`Approving museum item: ${itemId}...`);
    
    const response = await api.post(`/api/v1/museum/items/${itemId}/approve`);
    
    if (response.data.success) {
      console.log(`✓ Museum item approved successfully`);
      return true;
    } else {
      console.error(`✗ Failed to approve museum item: ${response.data.message}`);
      return false;
    }
  } catch (error) {
    console.error(`✗ Error approving museum item:`, error.response?.data || error.message);
    return false;
  }
}

// Helper function to verify museum entries
async function verifyMuseumEntries() {
  try {
    console.log('\nVerifying museum entries...');
    
    const response = await api.get('/api/v1/museum/entries', {
      params: {
        page: 1,
        limit: 10,
        status: 'approved'
      }
    });
    
    if (response.data.success) {
      const entries = response.data.data;
      console.log(`\n✓ Found ${entries.length} museum entries:`);
      
      entries.forEach((entry, index) => {
        console.log(`\n${index + 1}. ${entry.display_title || entry.title}`);
        console.log(`   Status: ${entry.status}`);
        console.log(`   Tags: ${entry.tags?.join(', ') || 'No tags'}`);
        console.log(`   Views: ${entry.view_count}, Likes: ${entry.like_count}`);
      });
      
      return entries;
    } else {
      console.error(`✗ Failed to get museum entries: ${response.data.message}`);
      return [];
    }
  } catch (error) {
    console.error(`✗ Error getting museum entries:`, error.response?.data || error.message);
    return [];
  }
}

// Main function
async function main() {
  console.log('=== OpenPenPal Museum Data Test Script ===');
  console.log(`API Base: ${API_BASE}`);
  console.log(`Using auth token: ${AUTH_TOKEN.substring(0, 20)}...`);
  console.log('');
  
  const createdItems = [];
  
  // Create letters and submit to museum
  for (const letterData of sampleLetters) {
    console.log(`\n--- Processing: "${letterData.title}" ---`);
    
    // Step 1: Create letter
    const letter = await createLetter(letterData);
    if (!letter) continue;
    
    // Step 2: Generate letter code
    const codeGenerated = await generateLetterCode(letter.id);
    if (!codeGenerated) {
      console.log('Skipping museum submission due to code generation failure');
      continue;
    }
    
    // Step 3: Submit to museum
    const museumItem = await submitToMuseum(letter, letterData);
    if (!museumItem) continue;
    
    createdItems.push(museumItem);
    
    // Step 4: Auto-approve (since we're admin)
    await approveMuseumItem(museumItem.id);
    
    // Add a small delay to avoid overwhelming the server
    await new Promise(resolve => setTimeout(resolve, 1000));
  }
  
  // Wait a bit for data to propagate
  console.log('\n\nWaiting for data to propagate...');
  await new Promise(resolve => setTimeout(resolve, 2000));
  
  // Verify museum entries
  const entries = await verifyMuseumEntries();
  
  // Summary
  console.log('\n\n=== Summary ===');
  console.log(`Total letters created: ${createdItems.length}`);
  console.log(`Museum entries found: ${entries.length}`);
  console.log('\nYou can now view these entries in the frontend at:');
  console.log('http://localhost:3000/museum');
  
  console.log('\n✅ Test completed successfully!');
}

// Run the script
main().catch(error => {
  console.error('\n❌ Script failed with error:', error);
  process.exit(1);
});