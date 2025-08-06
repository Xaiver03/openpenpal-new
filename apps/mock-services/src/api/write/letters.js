/**
 * 写信服务 API
 * 处理信件的创建、查询、更新等操作
 */

import { v4 as uuidv4 } from 'uuid';
import { createLogger } from '../../utils/logger.js';

const logger = createLogger('write-api');

// Mock 信件数据存储
let letters = [
  {
    id: 'letter_001',
    title: '给远方朋友的第一封信',
    content: '你好，这是我写给你的第一封信...',
    senderId: 'user_001',
    senderInfo: {
      username: 'alice',
      schoolCode: 'PKU',
      fullName: '爱丽丝'
    },
    receiverHint: '北京大学计算机系的朋友',
    receiverId: null,
    status: 'pending', // pending, matched, generated, printed, delivered, completed
    privacy: 'public',
    tags: ['友谊', '第一次'],
    metadata: {
      wordCount: 156,
      estimatedReadTime: 2,
      theme: 'friendship'
    },
    createdAt: new Date('2024-01-15T10:30:00Z').toISOString(),
    updatedAt: new Date('2024-01-15T10:30:00Z').toISOString()
  },
  {
    id: 'letter_002',
    title: '关于大学生活的思考',
    content: '最近在思考大学生活的意义...',
    senderId: 'user_002',
    senderInfo: {
      username: 'bob',
      schoolCode: 'THU',
      fullName: '鲍勃'
    },
    receiverHint: '清华大学物理系的同学',
    receiverId: 'user_001',
    status: 'matched',
    privacy: 'public',
    tags: ['思考', '大学生活'],
    metadata: {
      wordCount: 234,
      estimatedReadTime: 3,
      theme: 'reflection'
    },
    createdAt: new Date('2024-01-16T14:20:00Z').toISOString(),
    updatedAt: new Date('2024-01-16T15:45:00Z').toISOString()
  }
];

/**
 * 创建新信件
 */
export async function createLetter(req, res) {
  try {
    const { title, content, receiverHint, privacy = 'public', tags = [] } = req.body;
    
    // 验证输入
    const errors = [];
    if (!title) errors.push({ field: 'title', message: '标题不能为空' });
    if (!content) errors.push({ field: 'content', message: '内容不能为空' });
    if (!receiverHint) errors.push({ field: 'receiverHint', message: '收件人提示不能为空' });
    
    if (errors.length > 0) {
      return res.validationError(errors);
    }
    
    // 创建新信件
    const newLetter = {
      id: uuidv4(),
      title,
      content,
      senderId: req.user.id,
      senderInfo: {
        username: req.user.username,
        schoolCode: req.user.schoolCode,
        fullName: req.user.profile.fullName
      },
      receiverHint,
      receiverId: null,
      status: 'pending',
      privacy,
      tags: Array.isArray(tags) ? tags : [tags],
      metadata: {
        wordCount: content.length,
        estimatedReadTime: Math.ceil(content.length / 200),
        theme: 'general'
      },
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString()
    };
    
    letters.push(newLetter);
    
    logger.info(`新信件创建成功: ${newLetter.id} by ${req.user.username}`);
    
    return res.success(newLetter, '信件创建成功');
    
  } catch (error) {
    logger.error('创建信件异常:', error);
    return res.error(500, '创建信件失败');
  }
}

/**
 * 获取信件列表
 */
export async function getLetters(req, res) {
  try {
    const { 
      page = 0, 
      limit = 20, 
      status, 
      privacy, 
      senderId,
      search 
    } = req.query;
    
    let filteredLetters = [...letters];
    
    // 根据用户权限过滤
    if (req.user.role !== 'super_admin' && req.user.role !== 'admin') {
      // 普通用户只能看到自己的信件和公开的信件
      filteredLetters = filteredLetters.filter(letter => 
        letter.senderId === req.user.id || 
        letter.receiverId === req.user.id ||
        letter.privacy === 'public'
      );
    }
    
    // 状态过滤
    if (status) {
      filteredLetters = filteredLetters.filter(letter => letter.status === status);
    }
    
    // 隐私设置过滤
    if (privacy) {
      filteredLetters = filteredLetters.filter(letter => letter.privacy === privacy);
    }
    
    // 发送者过滤
    if (senderId) {
      filteredLetters = filteredLetters.filter(letter => letter.senderId === senderId);
    }
    
    // 搜索过滤
    if (search) {
      const searchLower = search.toLowerCase();
      filteredLetters = filteredLetters.filter(letter => 
        letter.title.toLowerCase().includes(searchLower) ||
        letter.content.toLowerCase().includes(searchLower) ||
        letter.tags.some(tag => tag.toLowerCase().includes(searchLower))
      );
    }
    
    // 分页处理
    const pageNum = parseInt(page, 10);
    const limitNum = parseInt(limit, 10);
    const startIndex = pageNum * limitNum;
    const endIndex = startIndex + limitNum;
    
    const paginatedLetters = filteredLetters.slice(startIndex, endIndex);
    
    const pagination = {
      page: pageNum,
      limit: limitNum,
      total: filteredLetters.length,
      pages: Math.ceil(filteredLetters.length / limitNum),
      hasNext: endIndex < filteredLetters.length,
      hasPrev: pageNum > 0
    };
    
    logger.info(`获取信件列表: ${paginatedLetters.length}/${filteredLetters.length} 条记录`);
    
    return res.paginated(paginatedLetters, pagination);
    
  } catch (error) {
    logger.error('获取信件列表异常:', error);
    return res.error(500, '获取信件列表失败');
  }
}

/**
 * 获取信件详情
 */
export async function getLetterById(req, res) {
  try {
    const { id } = req.params;
    
    const letter = letters.find(l => l.id === id);
    if (!letter) {
      return res.error(404, '信件不存在');
    }
    
    // 权限检查
    if (req.user.role !== 'super_admin' && req.user.role !== 'admin') {
      if (letter.senderId !== req.user.id && 
          letter.receiverId !== req.user.id && 
          letter.privacy !== 'public') {
        return res.error(403, '无权查看此信件');
      }
    }
    
    logger.info(`获取信件详情: ${id} by ${req.user.username}`);
    
    return res.success(letter, '获取信件详情成功');
    
  } catch (error) {
    logger.error('获取信件详情异常:', error);
    return res.error(500, '获取信件详情失败');
  }
}

/**
 * 更新信件状态
 */
export async function updateLetterStatus(req, res) {
  try {
    const { id } = req.params;
    const { status, adminNote } = req.body;
    
    const letter = letters.find(l => l.id === id);
    if (!letter) {
      return res.error(404, '信件不存在');
    }
    
    // 权限检查 - 只有管理员或信件所有者可以更新状态
    if (req.user.role !== 'super_admin' && 
        req.user.role !== 'admin' && 
        letter.senderId !== req.user.id) {
      return res.error(403, '无权更新此信件状态');
    }
    
    // 验证状态值
    const validStatuses = ['pending', 'matched', 'generated', 'printed', 'delivered', 'completed'];
    if (!validStatuses.includes(status)) {
      return res.validationError([
        { field: 'status', message: `状态值必须是: ${validStatuses.join(', ')}` }
      ]);
    }
    
    // 更新信件
    letter.status = status;
    letter.updatedAt = new Date().toISOString();
    
    if (adminNote) {
      letter.adminNote = adminNote;
    }
    
    logger.info(`信件状态更新: ${id} -> ${status} by ${req.user.username}`);
    
    return res.success(letter, '信件状态更新成功');
    
  } catch (error) {
    logger.error('更新信件状态异常:', error);
    return res.error(500, '更新信件状态失败');
  }
}

/**
 * 删除信件
 */
export async function deleteLetter(req, res) {
  try {
    const { id } = req.params;
    
    const letterIndex = letters.findIndex(l => l.id === id);
    if (letterIndex === -1) {
      return res.error(404, '信件不存在');
    }
    
    const letter = letters[letterIndex];
    
    // 权限检查 - 只有管理员或信件所有者可以删除
    if (req.user.role !== 'super_admin' && 
        req.user.role !== 'admin' && 
        letter.senderId !== req.user.id) {
      return res.error(403, '无权删除此信件');
    }
    
    // 只有待处理状态的信件可以删除
    if (letter.status !== 'pending') {
      return res.error(400, '只能删除待处理状态的信件');
    }
    
    letters.splice(letterIndex, 1);
    
    logger.info(`信件删除: ${id} by ${req.user.username}`);
    
    return res.success(null, '信件删除成功');
    
  } catch (error) {
    logger.error('删除信件异常:', error);
    return res.error(500, '删除信件失败');
  }
}

/**
 * 获取信件统计信息
 */
export async function getLetterStats(req, res) {
  try {
    // 根据用户权限过滤统计数据
    let userLetters = letters;
    if (req.user.role !== 'super_admin' && req.user.role !== 'admin') {
      userLetters = letters.filter(letter => 
        letter.senderId === req.user.id || letter.receiverId === req.user.id
      );
    }
    
    const stats = {
      total: userLetters.length,
      byStatus: {
        pending: userLetters.filter(l => l.status === 'pending').length,
        matched: userLetters.filter(l => l.status === 'matched').length,
        generated: userLetters.filter(l => l.status === 'generated').length,
        printed: userLetters.filter(l => l.status === 'printed').length,
        delivered: userLetters.filter(l => l.status === 'delivered').length,
        completed: userLetters.filter(l => l.status === 'completed').length
      },
      byPrivacy: {
        public: userLetters.filter(l => l.privacy === 'public').length,
        private: userLetters.filter(l => l.privacy === 'private').length
      },
      sent: userLetters.filter(l => l.senderId === req.user.id).length,
      received: userLetters.filter(l => l.receiverId === req.user.id).length
    };
    
    logger.info(`获取信件统计: ${req.user.username}`);
    
    return res.success(stats, '获取统计信息成功');
    
  } catch (error) {
    logger.error('获取信件统计异常:', error);
    return res.error(500, '获取统计信息失败');
  }
}

/**
 * 获取公开信件列表 - 无需认证
 * 用于广场页面展示
 */
export async function getPublicLetters(req, res) {
  try {
    const { 
      limit = 20, 
      sort_by = 'created_at',
      sort_order = 'desc',
      style
    } = req.query;
    
    // 扩展Mock数据，包含广场页面需要的样式分类
    const extendedLetters = [
      ...letters,
      {
        id: 'plaza-letter-1',
        title: '写给三年后的自己',
        content: '亲爱的未来的我，当你读到这封信的时候，希望你已经成为了更好的自己。还记得现在的我吗？那个在图书馆里挥汗如雨的学生，那个为了一道数学题而熬夜到凌晨的少年。我知道路还很长，但我相信，只要坚持下去，总会到达想要的地方...',
        senderId: 'plaza-user-1',
        senderInfo: {
          username: 'future_dreamer',
          schoolCode: 'PKU',
          fullName: '匿名作者'
        },
        style: 'future',
        privacy: 'public',
        status: 'completed',
        tags: ['成长', '梦想', '大学生活'],
        createdAt: '2024-01-20T10:00:00Z',
        updatedAt: '2024-01-20T10:00:00Z'
      },
      {
        id: 'plaza-letter-2',
        title: '致正在迷茫的你',
        content: '如果你正在经历人生的低谷，请记住这只是暂时的。每个人都会有迷茫的时候，这是成长路上的必经之路。不要害怕迷茫，因为只有经历过黑暗，我们才能更珍惜光明。愿你在迷雾中找到前进的方向，愿你的心永远充满希望...',
        senderId: 'plaza-user-2',
        senderInfo: {
          username: 'warm_messenger',
          schoolCode: 'THU',
          fullName: '温暖使者'
        },
        style: 'warm',
        privacy: 'public',
        status: 'completed',
        tags: ['鼓励', '治愈', '心理健康'],
        createdAt: '2024-01-19T14:30:00Z',
        updatedAt: '2024-01-19T14:30:00Z'
      },
      {
        id: 'plaza-letter-3',
        title: '一个关于友谊的故事',
        content: '我想和你分享一个关于友谊的故事，这个故事改变了我对友情的理解。那是一个秋天的下午，我坐在宿舍里感到孤独，突然收到了一个陌生人的来信...',
        senderId: 'plaza-user-3',
        senderInfo: {
          username: 'story_teller',
          schoolCode: 'BNU',
          fullName: '故事讲述者'
        },
        style: 'story',
        privacy: 'public',
        status: 'completed',
        tags: ['友谊', '青春', '回忆'],
        createdAt: '2024-01-18T09:15:00Z',
        updatedAt: '2024-01-18T09:15:00Z'
      },
      {
        id: 'plaza-letter-4',
        title: '漂流到远方的思念',
        content: '这封信将随风漂流到某个角落，希望能遇到同样思念远方的你。也许我们从未谋面，但在这个瞬间，我们的心是相通的...',
        senderId: 'plaza-user-4',
        senderInfo: {
          username: 'wanderer',
          schoolCode: 'FDU',
          fullName: '漂流者'
        },
        style: 'drift',
        privacy: 'public',
        status: 'completed',
        tags: ['思念', '漂流', '相遇'],
        createdAt: '2024-01-17T16:45:00Z',
        updatedAt: '2024-01-17T16:45:00Z'
      }
    ];
    
    // 只返回公开的信件
    let publicLetters = extendedLetters.filter(letter => letter.privacy === 'public');
    
    // 按样式过滤
    if (style && style !== 'all') {
      publicLetters = publicLetters.filter(letter => letter.style === style);
    }
    
    // 排序
    publicLetters.sort((a, b) => {
      const aValue = a[sort_by] || a.createdAt;
      const bValue = b[sort_by] || b.createdAt;
      
      if (sort_order === 'desc') {
        return new Date(bValue) - new Date(aValue);
      } else {
        return new Date(aValue) - new Date(bValue);
      }
    });
    
    // 限制数量
    const limitNum = parseInt(limit, 10) || 20;
    const limitedLetters = publicLetters.slice(0, limitNum);
    
    // 转换为前端需要的格式
    const formattedLetters = limitedLetters.map(letter => ({
      id: letter.id,
      title: letter.title,
      content: letter.content,
      user: {
        nickname: letter.senderInfo?.fullName || '匿名作者',
        avatar: `/images/user-${letter.id.split('-').pop()}.png`
      },
      style: letter.style || 'story',
      created_at: letter.createdAt,
      is_public: true
    }));
    
    logger.info(`获取公开信件列表: ${formattedLetters.length} 条记录 (style: ${style || 'all'})`);
    
    return res.success({
      data: formattedLetters,
      total: publicLetters.length,
      limit: limitNum
    }, '获取公开信件成功');
    
  } catch (error) {
    logger.error('获取公开信件异常:', error);
    return res.error(500, '获取公开信件失败');
  }
}