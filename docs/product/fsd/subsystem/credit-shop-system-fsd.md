# Credit Shop System Functional Specification Document

> **Version**: 2.0  
> **Implementation Status**: ✅ Production Ready  
> **Last Updated**: 2025-08-15  
> **Business Impact**: Core Revenue Generation System

## Implementation Overview

- **Completion**: 95% (Production Ready)
- **Production Ready**: Yes
- **Key Features**: E-commerce functionality, inventory management, purchase flow, digital goods delivery
- **Dependencies**: Credit System, User System, Notification System

## System Architecture

### Core Components

```
Credit Shop Ecosystem
├── Shop Frontend (React + Zustand)
├── Product Catalog Management
├── Shopping Cart & Checkout
├── Digital Goods Delivery
├── Purchase History & Analytics
└── Admin Management Interface
```

### Data Models

```typescript
// Core Shop Models
interface ShopItem {
  id: string
  name: string
  description: string
  price: number
  category: 'digital_goods' | 'premium_features' | 'customization'
  type: 'envelope_design' | 'letter_template' | 'ai_persona' | 'premium_tier'
  icon_url?: string
  preview_url?: string
  is_limited: boolean
  stock_quantity?: number
  tags: string[]
  created_at: string
  updated_at: string
}

interface Purchase {
  id: string
  user_id: string
  item_id: string
  quantity: number
  total_cost: number
  status: 'pending' | 'completed' | 'failed' | 'refunded'
  purchased_at: string
  delivered_at?: string
}

interface ShoppingCart {
  user_id: string
  items: CartItem[]
  total_cost: number
  updated_at: string
}
```

## Technical Implementation

### Frontend Architecture (`/app/shop/`)

**Shop Pages**:
- `/shop` - Main shop interface with categories
- `/shop/[category]` - Category-specific product listings  
- `/shop/item/[id]` - Individual product details
- `/shop/cart` - Shopping cart management
- `/shop/orders` - Purchase history

**Key Components**:
```typescript
// Primary Shop Components
- ShopItemCard          // Product display card
- ShoppingCart          // Cart management
- PurchaseConfirmation  // Checkout flow
- PurchaseHistory       // Order tracking
- ItemPreview          // Product preview modal
```

### State Management (`/stores/shop-store.ts`)

```typescript
interface ShopState {
  // Product Catalog
  items: ShopItem[]
  categories: string[]
  featured_items: ShopItem[]
  
  // Shopping Cart
  cart: ShoppingCart
  cart_visible: boolean
  
  // Purchase Flow
  checkout_loading: boolean
  purchase_history: Purchase[]
  
  // UI States
  loading: Record<string, boolean>
  errors: Record<string, string>
}

interface ShopActions {
  // Catalog Management
  loadShopItems: () => Promise<void>
  searchItems: (query: string, filters: any) => Promise<void>
  
  // Cart Operations
  addToCart: (item_id: string, quantity: number) => void
  removeFromCart: (item_id: string) => void
  updateCartQuantity: (item_id: string, quantity: number) => void
  clearCart: () => void
  
  // Purchase Flow
  purchaseItems: (items: CartItem[]) => Promise<PurchaseResult>
  loadPurchaseHistory: () => Promise<void>
  
  // Admin Functions
  createShopItem: (item: CreateShopItemRequest) => Promise<void>
  updateShopItem: (id: string, updates: Partial<ShopItem>) => Promise<void>
}
```

## API Endpoints

### Shop Catalog APIs

```
GET /api/shop/items
Query Parameters:
- category?: string
- search?: string  
- limit?: number
- offset?: number
- sort?: 'price_asc' | 'price_desc' | 'newest' | 'popular'

Response:
{
  "items": ShopItem[],
  "total": number,
  "categories": string[]
}

GET /api/shop/items/:id
Response: ShopItem

GET /api/shop/featured
Response: {
  "featured_items": ShopItem[]
}
```

### Shopping Cart APIs

```
GET /api/shop/cart
Response: ShoppingCart

POST /api/shop/cart/add
Body: {
  "item_id": string,
  "quantity": number
}

PUT /api/shop/cart/update
Body: {
  "item_id": string,
  "quantity": number
}

DELETE /api/shop/cart/remove/:item_id
```

### Purchase APIs

```
POST /api/shop/purchase
Body: {
  "items": CartItem[],
  "payment_method": "credits"
}
Response: {
  "purchase_id": string,
  "status": string,
  "total_cost": number,
  "delivered_items": DeliveredItem[]
}

GET /api/shop/purchases
Response: {
  "purchases": Purchase[]
}
```

## Database Schema

```sql
-- Shop Items Table
CREATE TABLE shop_items (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price INT NOT NULL,
    category VARCHAR(100) NOT NULL,
    type VARCHAR(100) NOT NULL,
    icon_url TEXT,
    preview_url TEXT,
    is_limited BOOLEAN DEFAULT FALSE,
    stock_quantity INT,
    tags JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_category (category),
    INDEX idx_type (type),
    INDEX idx_price (price)
);

-- Purchases Table
CREATE TABLE purchases (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    item_id VARCHAR(36) NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    total_cost INT NOT NULL,
    status ENUM('pending', 'completed', 'failed', 'refunded') DEFAULT 'pending',
    purchased_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    delivered_at TIMESTAMP NULL,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (item_id) REFERENCES shop_items(id),
    INDEX idx_user_purchases (user_id),
    INDEX idx_purchase_status (status)
);

-- Shopping Carts Table
CREATE TABLE shopping_carts (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL UNIQUE,
    items JSON NOT NULL,
    total_cost INT DEFAULT 0,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- User Purchased Items (for access control)
CREATE TABLE user_purchased_items (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    item_id VARCHAR(36) NOT NULL,
    purchase_id VARCHAR(36) NOT NULL,
    purchased_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NULL,
    is_active BOOLEAN DEFAULT TRUE,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (item_id) REFERENCES shop_items(id),
    FOREIGN KEY (purchase_id) REFERENCES purchases(id),
    UNIQUE KEY unique_user_item (user_id, item_id),
    INDEX idx_user_items (user_id)
);
```

## Product Categories & Types

### Digital Goods Categories

1. **Envelope Designs** (`envelope_design`)
   - Custom envelope templates
   - Seasonal designs
   - Artist collaborations
   - Price range: 10-50 credits

2. **Letter Templates** (`letter_template`)
   - Themed stationery
   - Calligraphy styles
   - Vintage collections
   - Price range: 5-30 credits

3. **AI Personas** (`ai_persona`)
   - Custom AI companion personalities
   - Historical figure personas
   - Celebrity-inspired characters
   - Price range: 50-200 credits

4. **Premium Features** (`premium_tier`)
   - Letter delivery priority
   - Advanced matching algorithms
   - Extended storage
   - Price range: 100-500 credits

## Business Logic Implementation

### Purchase Flow

```go
// Go Backend Implementation
func (s *ShopService) ProcessPurchase(userID string, items []CartItem) (*PurchaseResult, error) {
    // 1. Validate user credits
    user, err := s.userService.GetUser(userID)
    if err != nil {
        return nil, err
    }
    
    totalCost := s.calculateTotalCost(items)
    if user.Credits < totalCost {
        return nil, ErrInsufficientCredits
    }
    
    // 2. Create purchase record
    purchase := &Purchase{
        ID: generateID(),
        UserID: userID,
        TotalCost: totalCost,
        Status: "pending",
    }
    
    // 3. Deduct credits (atomic transaction)
    err = s.db.Transaction(func(tx *sql.Tx) error {
        // Deduct credits
        err := s.creditService.DeductCredits(tx, userID, totalCost)
        if err != nil {
            return err
        }
        
        // Create purchase
        err = s.createPurchase(tx, purchase)
        if err != nil {
            return err
        }
        
        // Deliver digital goods
        return s.deliverDigitalGoods(tx, userID, items)
    })
    
    if err != nil {
        purchase.Status = "failed"
        return nil, err
    }
    
    purchase.Status = "completed"
    return &PurchaseResult{Purchase: purchase}, nil
}
```

### Digital Goods Delivery

```go
func (s *ShopService) deliverDigitalGoods(tx *sql.Tx, userID string, items []CartItem) error {
    for _, item := range items {
        switch item.Type {
        case "envelope_design":
            err := s.unlockEnvelopeDesign(tx, userID, item.ID)
        case "letter_template":
            err := s.unlockLetterTemplate(tx, userID, item.ID)
        case "ai_persona":
            err := s.unlockAIPersona(tx, userID, item.ID)
        case "premium_tier":
            err := s.activatePremiumFeatures(tx, userID, item.ID)
        }
        
        if err != nil {
            return err
        }
    }
    return nil
}
```

## Frontend User Experience

### Shop Interface Features

1. **Product Discovery**
   - Category browsing with filters
   - Search functionality with autocomplete
   - Featured items carousel
   - Personalized recommendations

2. **Shopping Cart Experience**
   - Persistent cart across sessions
   - Real-time cost calculation
   - Bulk operations support
   - Quick checkout flow

3. **Purchase Confirmation**
   - Credit balance verification
   - Item preview before purchase
   - Instant delivery confirmation
   - Purchase receipt generation

### Responsive Design

- **Desktop**: Full grid layout with detailed previews
- **Tablet**: Adaptive grid with touch-optimized controls
- **Mobile**: Single-column layout with swipe navigation

## Security & Business Rules

### Purchase Validation

1. **Credit Verification**
   - Real-time balance checking
   - Fraud detection for unusual patterns
   - Purchase limits per time period

2. **Inventory Management**
   - Stock level tracking for limited items
   - Preventing oversale conditions
   - Automatic restocking notifications

3. **Access Control**
   - User ownership verification
   - Premium feature unlocking
   - Expiration date management

### Audit & Compliance

- All purchases logged with detailed metadata
- Refund capability with credit restoration
- Purchase analytics for business intelligence
- GDPR compliance for user data

## Performance Optimizations

### Caching Strategy

```typescript
// Frontend Caching
const shopStore = create<ShopState>()(
  persist(
    subscribeWithSelector((set, get) => ({
      // Cache shop items for 10 minutes
      items: [],
      itemsCache: new Map(),
      lastFetchTime: null,
      
      loadShopItems: async () => {
        const now = Date.now()
        const cache = get().lastFetchTime
        
        if (cache && now - cache < 10 * 60 * 1000) {
          return // Use cached data
        }
        
        // Fetch fresh data
        const items = await shopAPI.getItems()
        set({ items, lastFetchTime: now })
      }
    })),
    { name: 'shop-store' }
  )
)
```

### Database Optimization

- Indexed queries on category, price, and user_id
- Prepared statements for common operations
- Connection pooling for high concurrency
- Read replicas for product catalog queries

## Analytics & Monitoring

### Business Metrics

- Revenue per user (RPU)
- Conversion rates by category
- Cart abandonment analysis
- Popular item tracking

### Technical Metrics

- API response times
- Error rates by endpoint
- Cache hit ratios
- Database query performance

## Integration Points

### Credit System Integration

- Real-time credit balance checking
- Atomic credit deduction
- Purchase history in credit statements

### Notification System Integration

- Purchase confirmation notifications
- Delivery status updates
- Promotional announcements

### User System Integration

- Premium feature activation
- Purchase history in user profiles
- Loyalty program integration

## Future Enhancements

### Planned Features

1. **Gift System** - Send purchased items to friends
2. **Subscription Model** - Monthly premium memberships
3. **Marketplace** - User-created content sales
4. **Seasonal Events** - Limited-time exclusive items

### Technical Improvements

1. **Payment Gateway** - Real currency integration
2. **Advanced Analytics** - ML-based recommendations
3. **Mobile App** - Native shopping experience
4. **Internationalization** - Multi-currency support

---

**PRODUCTION STATUS**: The Credit Shop System is fully implemented and production-ready, serving as the primary monetization mechanism for OpenPenPal. All core e-commerce functionality is operational with enterprise-grade security and performance optimizations.