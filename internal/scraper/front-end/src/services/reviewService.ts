import { Review, ReviewStats, Platform, Department } from '../types/reviews';
import { v4 as uuidv4 } from 'uuid';

// Mock data for the review service
const mockReviews: Review[] = [
  {
    id: uuidv4(),
    title: 'Great product, highly recommend!',
    content: 'This tool has completely transformed how we handle customer feedback. The insights are invaluable.',
    rating: 5,
    date: '2025-04-20T15:32:00Z',
    platform: 'G2',
    sentiment: 'Positive',
    department: 'Product',
    author: 'John Doe',
    url: 'https://www.g2.com/products/review/123456',
    highlights: ['Easy to use', 'Great insights', 'Excellent support'],
    tags: ['analytics', 'customer feedback', 'user-friendly'],
    isProcessed: true
  },
  {
    id: uuidv4(),
    title: 'Could use some improvements',
    content: 'The tool is good overall but has a steep learning curve. Documentation could be better.',
    rating: 3,
    date: '2025-04-18T09:45:00Z',
    platform: 'Trustpilot',
    sentiment: 'Neutral',
    department: 'Support',
    author: 'Jane Smith',
    url: 'https://www.trustpilot.com/review/product/789012',
    highlights: ['Powerful features', 'Complex interface', 'Learning curve'],
    tags: ['documentation', 'usability', 'features'],
    isProcessed: true
  },
  {
    id: uuidv4(),
    title: 'Disappointing experience',
    content: 'The app keeps crashing and support has been unresponsive for days.',
    rating: 1,
    date: '2025-04-17T14:22:00Z',
    platform: 'App Store',
    sentiment: 'Negative',
    department: 'Engineering',
    author: 'Mike Johnson',
    url: 'https://apps.apple.com/us/app/product/id345678',
    highlights: ['Crashes frequently', 'Poor support response'],
    tags: ['bugs', 'reliability', 'support'],
    isProcessed: true
  },
  {
    id: uuidv4(),
    title: 'Game-changing tool',
    content: 'This has dramatically improved our workflow. We are seeing a 40% increase in productivity.',
    rating: 5,
    date: '2025-04-15T18:12:00Z',
    platform: 'G2',
    sentiment: 'Positive',
    department: 'Product',
    author: 'Sarah Williams',
    url: 'https://www.g2.com/products/review/567890',
    highlights: ['Productivity boost', 'Time-saving', 'Intuitive'],
    tags: ['efficiency', 'workflow', 'productivity'],
    isProcessed: true
  },
  {
    id: uuidv4(),
    title: 'Not worth the price',
    content: 'Too expensive for what it offers. There are better alternatives at half the price.',
    rating: 2,
    date: '2025-04-14T10:05:00Z',
    platform: 'Trustpilot',
    sentiment: 'Negative',
    department: 'Sales',
    author: 'Robert Brown',
    url: 'https://www.trustpilot.com/review/product/234567',
    highlights: ['Overpriced', 'Better alternatives'],
    tags: ['pricing', 'value', 'competition'],
    isProcessed: true
  },
  {
    id: uuidv4(),
    title: 'Solid product with room for improvement',
    content: 'Good feature set but the UI could be more intuitive. Support team is very helpful.',
    rating: 4,
    date: '2025-04-10T13:45:00Z',
    platform: 'Reddit',
    sentiment: 'Positive',
    department: 'Support',
    author: 'u/techreviewer42',
    url: 'https://www.reddit.com/r/SoftwareReviews/comments/abc123',
    highlights: ['Good features', 'Helpful support', 'UI needs work'],
    tags: ['user interface', 'features', 'support'],
    isProcessed: true
  },
  {
    id: uuidv4(),
    title: 'Mixed feelings about this tool',
    content: 'Some features are great, others feel half-baked. The analytics are impressive but exporting is limited.',
    rating: 3,
    date: '2025-04-08T09:30:00Z',
    platform: 'Google Play',
    sentiment: 'Neutral',
    department: 'Product',
    author: 'Emily Davis',
    url: 'https://play.google.com/store/apps/details?id=com.product',
    highlights: ['Good analytics', 'Limited exports', 'Inconsistent quality'],
    tags: ['analytics', 'export', 'features'],
    isProcessed: true
  },
];

// Mock stats based on the mock reviews
const calculateMockStats = (): ReviewStats => {
  const positiveCount = mockReviews.filter(r => r.sentiment === 'Positive').length;
  const neutralCount = mockReviews.filter(r => r.sentiment === 'Neutral').length;
  const negativeCount = mockReviews.filter(r => r.sentiment === 'Negative').length;
  
  const totalRating = mockReviews.reduce((sum, review) => sum + (review.rating || 0), 0);
  const reviewsWithRating = mockReviews.filter(r => r.rating !== undefined).length;
  
  const byPlatform: Record<Platform, number> = {
    'G2': 0,
    'App Store': 0,
    'Google Play': 0,
    'Trustpilot': 0,
    'Reddit': 0,
    'Twitter': 0,
    'Other': 0
  };
  
  const byDepartment: Record<Department, number> = {
    'Product': 0,
    'Support': 0,
    'Sales': 0,
    'Marketing': 0,
    'Engineering': 0,
    'General': 0
  };
  
  mockReviews.forEach(review => {
    byPlatform[review.platform]++;
    byDepartment[review.department]++;
  });
  
  return {
    totalReviews: mockReviews.length,
    positiveCount,
    neutralCount,
    negativeCount,
    averageRating: reviewsWithRating > 0 ? totalRating / reviewsWithRating : 0,
    byPlatform,
    byDepartment,
    recentTrend: positiveCount > negativeCount ? 'up' : negativeCount > positiveCount ? 'down' : 'stable'
  };
};

const mockStats: ReviewStats = calculateMockStats();

// Simulate API delay
const delay = (ms: number) => new Promise(resolve => setTimeout(resolve, ms));

export const reviewService = {
  // Get all reviews with optional filtering
  getReviews: async (filters?: {
    platform?: Platform,
    department?: Department,
    sentiment?: string,
    searchTerm?: string
  }): Promise<Review[]> => {
    await delay(800); // Simulate API delay
    
    if (!filters) return mockReviews;
    
    return mockReviews.filter(review => {
      if (filters.platform && review.platform !== filters.platform) return false;
      if (filters.department && review.department !== filters.department) return false;
      if (filters.sentiment && review.sentiment !== filters.sentiment) return false;
      if (filters.searchTerm) {
        const term = filters.searchTerm.toLowerCase();
        return (
          review.title.toLowerCase().includes(term) ||
          review.content.toLowerCase().includes(term) ||
          review.author?.toLowerCase().includes(term) ||
          review.tags?.some(tag => tag.toLowerCase().includes(term)) ||
          false
        );
      }
      return true;
    });
  },
  
  // Get a single review by ID
  getReviewById: async (id: string): Promise<Review | null> => {
    await delay(500); // Simulate API delay
    return mockReviews.find(review => review.id === id) || null;
  },
  
  // Update a review's processing status
  updateReviewStatus: async (id: string, isProcessed: boolean): Promise<Review | null> => {
    await delay(700); // Simulate API delay
    const index = mockReviews.findIndex(review => review.id === id);
    if (index !== -1) {
      mockReviews[index] = { ...mockReviews[index], isProcessed };
      return mockReviews[index];
    }
    return null;
  },
  
  // Get statistics about reviews
  getStats: async (): Promise<ReviewStats> => {
    await delay(1000); // Simulate API delay
    return mockStats;
  },
  
  // Trigger a new scraping job
  triggerScraping: async (platforms: Platform[]): Promise<{ jobId: string, message: string }> => {
    await delay(1500); // Simulate API delay
    return {
      jobId: uuidv4(),
      message: `Scraping job started for platforms: ${platforms.join(', ')}`
    };
  }
};

export default reviewService;