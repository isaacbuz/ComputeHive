import React from 'react';
import { Grid, GridProps } from '@mui/material';
import { useResponsive } from '../../hooks/useResponsive';

interface ResponsiveGridProps extends Omit<GridProps, 'item' | 'container'> {
  children: React.ReactNode;
  columns?: {
    xs?: number;
    sm?: number;
    md?: number;
    lg?: number;
    xl?: number;
  };
}

interface ResponsiveGridItemProps extends Omit<GridProps, 'item' | 'container'> {
  children: React.ReactNode;
  size?: {
    xs?: number;
    sm?: number;
    md?: number;
    lg?: number;
    xl?: number;
  };
}

export const ResponsiveGrid: React.FC<ResponsiveGridProps> = ({
  children,
  columns = { xs: 4, sm: 8, md: 12, lg: 12, xl: 12 },
  spacing = 2,
  ...props
}) => {
  const { breakpoint } = useResponsive();
  
  // Get current columns based on breakpoint
  const currentColumns = columns[breakpoint] || 12;

  return (
    <Grid container spacing={spacing} columns={currentColumns} {...props}>
      {children}
    </Grid>
  );
};

export const ResponsiveGridItem: React.FC<ResponsiveGridItemProps> = ({
  children,
  size = { xs: 12, sm: 6, md: 4, lg: 3, xl: 3 },
  ...props
}) => {
  return (
    <Grid
      item
      xs={size.xs}
      sm={size.sm}
      md={size.md}
      lg={size.lg}
      xl={size.xl}
      {...props}
    >
      {children}
    </Grid>
  );
};

// Preset grid layouts
export const GridLayouts = {
  // Single column on mobile, 2 on tablet, 3 on desktop
  cards: {
    xs: 12,
    sm: 6,
    md: 4,
    lg: 4,
    xl: 3,
  },
  
  // Full width on mobile, half on tablet and up
  twoColumn: {
    xs: 12,
    sm: 6,
    md: 6,
    lg: 6,
    xl: 6,
  },
  
  // Full width on mobile/tablet, third on desktop
  threeColumn: {
    xs: 12,
    sm: 12,
    md: 4,
    lg: 4,
    xl: 4,
  },
  
  // Sidebar layout: 1/3 sidebar, 2/3 content
  sidebar: {
    xs: 12,
    sm: 12,
    md: 4,
    lg: 3,
    xl: 3,
  },
  
  content: {
    xs: 12,
    sm: 12,
    md: 8,
    lg: 9,
    xl: 9,
  },
  
  // Dashboard widgets
  widget: {
    xs: 12,
    sm: 6,
    md: 6,
    lg: 3,
    xl: 3,
  },
  
  // Form layouts
  formFull: {
    xs: 12,
    sm: 12,
    md: 12,
    lg: 12,
    xl: 12,
  },
  
  formHalf: {
    xs: 12,
    sm: 12,
    md: 6,
    lg: 6,
    xl: 6,
  },
};

// Responsive Masonry Grid for variable height items
interface ResponsiveMasonryProps {
  children: React.ReactNode;
  columns?: {
    xs?: number;
    sm?: number;
    md?: number;
    lg?: number;
    xl?: number;
  };
  spacing?: number;
}

export const ResponsiveMasonry: React.FC<ResponsiveMasonryProps> = ({
  children,
  columns = { xs: 1, sm: 2, md: 3, lg: 4, xl: 4 },
  spacing = 2,
}) => {
  const { breakpoint } = useResponsive();
  const columnCount = columns[breakpoint] || 1;

  return (
    <div
      style={{
        columnCount,
        columnGap: spacing * 8,
      }}
    >
      {React.Children.map(children, (child, index) => (
        <div
          key={index}
          style={{
            breakInside: 'avoid',
            marginBottom: spacing * 8,
          }}
        >
          {child}
        </div>
      ))}
    </div>
  );
};
