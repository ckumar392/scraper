import React, { useState, useEffect } from 'react';
import { 
  Container, Typography, Box,  
  TextField, InputAdornment, MenuItem,
  Select, FormControl, InputLabel, SelectChangeEvent,
  Chip, CircularProgress, Pagination, Grid
} from '@mui/material';
import SearchIcon from '@mui/icons-material/Search';
import FilterListIcon from '@mui/icons-material/FilterList';
import { motion, AnimatePresence } from 'framer-motion';
import { styled } from '@mui/material/styles';
import ReviewCard from '../components/ReviewCard';
import AnimatedCard from '../components/AnimatedCard';
import AnimatedButton from '../components/AnimatedButton';
import { Review, Platform, Department, Sentiment } from '../types/reviews';
import { reviewService } from '../services/reviewService';
import { colors } from '../theme/theme';

// Other styled components
const HeaderBox = styled(Box)(({ theme }) => ({
  marginBottom: theme.spacing(4),
  display: 'flex',
  justifyContent: 'space-between',
  alignItems: 'center',
  [theme.breakpoints.down('sm')]: {
    flexDirection: 'column',
    alignItems: 'stretch',
    gap: theme.spacing(2),
  }
}));

const GradientText = styled(Typography)(({ theme }) => ({
  fontWeight: 'bold',
  backgroundImage: `linear-gradient(135deg, ${colors.primary} 0%, ${colors.accent2} 100%)`,
  backgroundClip: 'text',
  WebkitBackgroundClip: 'text',
  color: 'transparent',
}));

const FilterContainer = styled(Box)(({ theme }) => ({
  padding: theme.spacing(3),
  marginBottom: theme.spacing(4),
  borderRadius: '16px',
  background: 'rgba(255, 255, 255, 0.8)',
  backdropFilter: 'blur(10px)',
  boxShadow: '0 8px 32px 0 rgba(0, 0, 0, 0.05)',
}));

const FilterChip = styled(Chip)<{ isactive: string }>(({ isactive, theme }) => ({
  margin: theme.spacing(0.5),
  backgroundColor: isactive === 'true' ? colors.primary : 'transparent',
  color: isactive === 'true' ? '#FFFFFF' : colors.darkGray,
  borderColor: isactive === 'true' ? 'transparent' : colors.primary,
  '&:hover': {
    backgroundColor: isactive === 'true' ? colors.primary : 'rgba(0, 114, 198, 0.08)',
  }
}));

const LoaderContainer = styled(Box)(({ theme }) => ({
  display: 'flex',
  justifyContent: 'center',
  alignItems: 'center',
  height: '400px'
}));

const PaginationContainer = styled(Box)(({ theme }) => ({
  display: 'flex',
  justifyContent: 'center',
  padding: theme.spacing(4, 0),
}));

const ReviewsPage: React.FC = () => {
  const [reviews, setReviews] = useState<Review[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [searchTerm, setSearchTerm] = useState<string>('');
  const [selectedPlatform, setSelectedPlatform] = useState<Platform | ''>('');
  const [selectedDepartment, setSelectedDepartment] = useState<Department | ''>('');
  const [selectedSentiment, setSelectedSentiment] = useState<string>('');
  const [currentPage, setCurrentPage] = useState<number>(1);
  const [reviewsPerPage] = useState<number>(6);

  const platforms: Platform[] = ['G2', 'App Store', 'Google Play', 'Trustpilot', 'Reddit', 'Twitter', 'Other'];
  const departments: Department[] = ['Product', 'Support', 'Sales', 'Marketing', 'Engineering', 'General'];
  const sentiments: Sentiment[] = ['Positive', 'Neutral', 'Negative'];

  // Fetch reviews on component load and when filters change
  useEffect(() => {
    const fetchReviews = async () => {
      setLoading(true);
      try {
        const filters: {
          platform?: Platform,
          department?: Department,
          sentiment?: string,
          searchTerm?: string
        } = {};

        if (selectedPlatform) filters.platform = selectedPlatform;
        if (selectedDepartment) filters.department = selectedDepartment;
        if (selectedSentiment) filters.sentiment = selectedSentiment;
        if (searchTerm) filters.searchTerm = searchTerm;

        const data = await reviewService.getReviews(Object.keys(filters).length > 0 ? filters : undefined);
        setReviews(data);
      } catch (error) {
        console.error('Error fetching reviews:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchReviews();
  }, [selectedPlatform, selectedDepartment, selectedSentiment, searchTerm]);

  // Handle filter changes
  const handlePlatformChange = (event: SelectChangeEvent<string>) => {
    setSelectedPlatform(event.target.value as Platform | '');
    setCurrentPage(1);
  };

  const handleDepartmentChange = (event: SelectChangeEvent<string>) => {
    setSelectedDepartment(event.target.value as Department | '');
    setCurrentPage(1);
  };

  const handleSentimentFilter = (sentiment: string) => {
    setSelectedSentiment(selectedSentiment === sentiment ? '' : sentiment);
    setCurrentPage(1);
  };

  const handleSearchChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setSearchTerm(event.target.value);
    setCurrentPage(1);
  };

  const handleReviewClick = (review: Review) => {
    console.log('Review clicked:', review);
    // Can be expanded to show a detailed view or modal
  };

  const handlePageChange = (_event: React.ChangeEvent<unknown>, value: number) => {
    setCurrentPage(value);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  // Pagination
  const indexOfLastReview = currentPage * reviewsPerPage;
  const indexOfFirstReview = indexOfLastReview - reviewsPerPage;
  const currentReviews = reviews.slice(indexOfFirstReview, indexOfLastReview);
  const totalPages = Math.ceil(reviews.length / reviewsPerPage);

  const clearFilters = () => {
    setSelectedPlatform('');
    setSelectedDepartment('');
    setSelectedSentiment('');
    setSearchTerm('');
    setCurrentPage(1);
  };

  return (
    <Container maxWidth="lg" sx={{ py: 4 }}>
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
      >
        <HeaderBox>
          <GradientText variant="h3">
            Reviews
          </GradientText>
        </HeaderBox>
      </motion.div>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.1, duration: 0.5 }}
      >
        <FilterContainer>
          <Box sx={{ display: 'flex', alignItems: 'center', mb: 3 }}>
            <FilterListIcon sx={{ mr: 1, color: colors.primary }} />
            <Typography variant="h6">Filters</Typography>
            <Box sx={{ flexGrow: 1 }} />
            <AnimatedButton
              variant="outlined"
              size="small"
              onClick={clearFilters}
            >
              Clear Filters
            </AnimatedButton>
          </Box>
          
          <Grid container spacing={3}>
            <Grid item xs={12} md={4}>
              <TextField
                fullWidth
                variant="outlined"
                placeholder="Search reviews..."
                value={searchTerm}
                onChange={handleSearchChange}
                InputProps={{
                  startAdornment: (
                    <InputAdornment position="start">
                      <SearchIcon />
                    </InputAdornment>
                  ),
                }}
              />
            </Grid>
            
            <Grid item xs={12} sm={6} md={4}>
              <FormControl fullWidth>
                <InputLabel id="platform-label">Platform</InputLabel>
                <Select
                  labelId="platform-label"
                  value={selectedPlatform}
                  label="Platform"
                  onChange={handlePlatformChange}
                >
                  <MenuItem value="">
                    <em>All Platforms</em>
                  </MenuItem>
                  {platforms.map((platform) => (
                    <MenuItem key={platform} value={platform}>{platform}</MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
            
            <Grid item xs={12} sm={6} md={4}>
              <FormControl fullWidth>
                <InputLabel id="department-label">Department</InputLabel>
                <Select
                  labelId="department-label"
                  value={selectedDepartment}
                  label="Department"
                  onChange={handleDepartmentChange}
                >
                  <MenuItem value="">
                    <em>All Departments</em>
                  </MenuItem>
                  {departments.map((department) => (
                    <MenuItem key={department} value={department}>{department}</MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
          </Grid>
          
          <Box sx={{ mt: 3 }}>
            <Typography variant="subtitle2" sx={{ mb: 1 }}>Sentiment:</Typography>
            <Box sx={{ display: 'flex', flexWrap: 'wrap' }}>
              {sentiments.map((sentiment) => (
                <motion.div
                  key={sentiment}
                  whileHover={{ scale: 1.05 }}
                  whileTap={{ scale: 0.95 }}
                >
                  <FilterChip
                    label={sentiment}
                    isactive={(selectedSentiment === sentiment).toString()}
                    variant={selectedSentiment === sentiment ? "filled" : "outlined"}
                    onClick={() => handleSentimentFilter(sentiment)}
                  />
                </motion.div>
              ))}
            </Box>
          </Box>
        </FilterContainer>
      </motion.div>

      {loading ? (
        <LoaderContainer>
          <CircularProgress size={60} thickness={4} sx={{ color: colors.accent2 }} />
        </LoaderContainer>
      ) : reviews.length > 0 ? (
        <>
          <AnimatePresence>
            <Grid container spacing={3}>
              {currentReviews.map((review, index) => (
                <Grid item xs={12} md={6} key={review.id}>
                  <Box>
                    <ReviewCard 
                      review={review} 
                      index={index} 
                      onClick={handleReviewClick}
                    />
                  </Box>
                </Grid>
              ))}
            </Grid>
          </AnimatePresence>
          
          {totalPages > 1 && (
            <PaginationContainer>
              <Pagination 
                count={totalPages} 
                page={currentPage}
                onChange={handlePageChange}
                color="primary"
                size="large"
                sx={{
                  '& .MuiPaginationItem-root': {
                    fontWeight: 'bold',
                  }
                }}
              />
            </PaginationContainer>
          )}
        </>
      ) : (
        <AnimatedCard sx={{ p: 4, textAlign: 'center' }}>
          <Typography variant="h6" gutterBottom>
            No reviews found
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Try adjusting your filters or search criteria
          </Typography>
        </AnimatedCard>
      )}
    </Container>
  );
};

export default ReviewsPage;