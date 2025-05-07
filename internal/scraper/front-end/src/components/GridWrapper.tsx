import React from 'react';
import { Grid as MuiGrid } from '@mui/material';

// Define a flexible type for Grid props
interface FlexibleGridProps {
  container?: boolean;
  item?: boolean;
  xs?: number | boolean;
  sm?: number | boolean;
  md?: number | boolean;
  lg?: number | boolean;
  xl?: number | boolean;
  spacing?: number;
  children?: React.ReactNode;
  sx?: any;
  component?: any;
  key?: string;
  [key: string]: any; // This allows any other props
}

// Create a wrapper component that bypasses type checking for Grid
export const Grid: React.FC<FlexibleGridProps> = (props) => {
  // @ts-ignore - Ignore TypeScript errors for Grid props
  return <MuiGrid {...props} />;
};

export default Grid;
