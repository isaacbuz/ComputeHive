import React, { useState } from 'react';
import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Card,
  CardContent,
  Typography,
  Box,
  Chip,
  IconButton,
  Menu,
  MenuItem,
  TablePagination,
  TextField,
  InputAdornment,
  Collapse,
  useTheme,
  alpha,
} from '@mui/material';
import {
  MoreVert as MoreIcon,
  Search as SearchIcon,
  KeyboardArrowDown as ExpandIcon,
  KeyboardArrowUp as CollapseIcon,
} from '@mui/icons-material';
import { useResponsive } from '../../hooks/useResponsive';
import { styled } from '@mui/material/styles';

export interface Column<T> {
  id: keyof T;
  label: string;
  minWidth?: number;
  align?: 'right' | 'left' | 'center';
  format?: (value: any) => React.ReactNode;
  mobileLabel?: string;
  hideOnMobile?: boolean;
  primary?: boolean;
}

interface ResponsiveDataTableProps<T> {
  columns: Column<T>[];
  rows: T[];
  onRowClick?: (row: T) => void;
  actions?: Array<{
    label: string;
    onClick: (row: T) => void;
    icon?: React.ReactNode;
  }>;
  searchable?: boolean;
  expandable?: boolean;
  renderExpandedContent?: (row: T) => React.ReactNode;
}

const StyledTableRow = styled(TableRow)(({ theme }) => ({
  '&:hover': {
    backgroundColor: alpha(theme.palette.primary.main, 0.05),
    cursor: 'pointer',
  },
}));

const MobileCard = styled(Card)(({ theme }) => ({
  marginBottom: theme.spacing(1),
  '&:last-child': {
    marginBottom: 0,
  },
  transition: 'all 0.2s ease',
  '&:hover': {
    transform: 'translateY(-2px)',
    boxShadow: theme.shadows[4],
  },
}));

export function ResponsiveDataTable<T extends { id: string | number }>({
  columns,
  rows,
  onRowClick,
  actions,
  searchable = true,
  expandable = false,
  renderExpandedContent,
}: ResponsiveDataTableProps<T>) {
  const { isMobile } = useResponsive();
  const theme = useTheme();
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10);
  const [searchQuery, setSearchQuery] = useState('');
  const [expandedRows, setExpandedRows] = useState<Set<string | number>>(new Set());
  const [anchorEl, setAnchorEl] = useState<{ [key: string]: HTMLElement | null }>({});

  const handleChangePage = (event: unknown, newPage: number) => {
    setPage(newPage);
  };

  const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
    setRowsPerPage(+event.target.value);
    setPage(0);
  };

  const handleMenuClick = (event: React.MouseEvent<HTMLElement>, rowId: string | number) => {
    event.stopPropagation();
    setAnchorEl({ ...anchorEl, [rowId]: event.currentTarget });
  };

  const handleMenuClose = (rowId: string | number) => {
    setAnchorEl({ ...anchorEl, [rowId]: null });
  };

  const toggleRowExpansion = (rowId: string | number) => {
    const newExpanded = new Set(expandedRows);
    if (newExpanded.has(rowId)) {
      newExpanded.delete(rowId);
    } else {
      newExpanded.add(rowId);
    }
    setExpandedRows(newExpanded);
  };

  // Filter rows based on search query
  const filteredRows = rows.filter((row) => {
    if (!searchQuery) return true;
    return columns.some((column) => {
      const value = row[column.id];
      return String(value).toLowerCase().includes(searchQuery.toLowerCase());
    });
  });

  // Paginate rows
  const paginatedRows = filteredRows.slice(
    page * rowsPerPage,
    page * rowsPerPage + rowsPerPage
  );

  // Get primary column for mobile display
  const primaryColumn = columns.find((col) => col.primary) || columns[0];

  // Desktop Table View
  const DesktopTable = () => (
    <Paper sx={{ width: '100%', overflow: 'hidden' }}>
      {searchable && (
        <Box sx={{ p: 2, borderBottom: 1, borderColor: 'divider' }}>
          <TextField
            fullWidth
            size="small"
            placeholder="Search..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">
                  <SearchIcon />
                </InputAdornment>
              ),
            }}
          />
        </Box>
      )}
      
      <TableContainer sx={{ maxHeight: 440 }}>
        <Table stickyHeader>
          <TableHead>
            <TableRow>
              {expandable && <TableCell />}
              {columns.map((column) => (
                <TableCell
                  key={String(column.id)}
                  align={column.align}
                  style={{ minWidth: column.minWidth }}
                >
                  {column.label}
                </TableCell>
              ))}
              {actions && <TableCell align="right">Actions</TableCell>}
            </TableRow>
          </TableHead>
          <TableBody>
            {paginatedRows.map((row) => (
              <React.Fragment key={row.id}>
                <StyledTableRow
                  onClick={() => onRowClick && onRowClick(row)}
                >
                  {expandable && (
                    <TableCell>
                      <IconButton
                        size="small"
                        onClick={(e) => {
                          e.stopPropagation();
                          toggleRowExpansion(row.id);
                        }}
                      >
                        {expandedRows.has(row.id) ? <CollapseIcon /> : <ExpandIcon />}
                      </IconButton>
                    </TableCell>
                  )}
                  {columns.map((column) => {
                    const value = row[column.id];
                    return (
                      <TableCell key={String(column.id)} align={column.align}>
                        {column.format ? column.format(value) : String(value)}
                      </TableCell>
                    );
                  })}
                  {actions && (
                    <TableCell align="right">
                      <IconButton onClick={(e) => handleMenuClick(e, row.id)}>
                        <MoreIcon />
                      </IconButton>
                      <Menu
                        anchorEl={anchorEl[row.id]}
                        open={Boolean(anchorEl[row.id])}
                        onClose={() => handleMenuClose(row.id)}
                      >
                        {actions.map((action, index) => (
                          <MenuItem
                            key={index}
                            onClick={() => {
                              action.onClick(row);
                              handleMenuClose(row.id);
                            }}
                          >
                            {action.icon && <Box sx={{ mr: 1 }}>{action.icon}</Box>}
                            {action.label}
                          </MenuItem>
                        ))}
                      </Menu>
                    </TableCell>
                  )}
                </StyledTableRow>
                {expandable && renderExpandedContent && (
                  <TableRow>
                    <TableCell colSpan={columns.length + (actions ? 2 : 1)}>
                      <Collapse in={expandedRows.has(row.id)}>
                        <Box sx={{ p: 2 }}>
                          {renderExpandedContent(row)}
                        </Box>
                      </Collapse>
                    </TableCell>
                  </TableRow>
                )}
              </React.Fragment>
            ))}
          </TableBody>
        </Table>
      </TableContainer>
      
      <TablePagination
        rowsPerPageOptions={[10, 25, 100]}
        component="div"
        count={filteredRows.length}
        rowsPerPage={rowsPerPage}
        page={page}
        onPageChange={handleChangePage}
        onRowsPerPageChange={handleChangeRowsPerPage}
      />
    </Paper>
  );

  // Mobile Card View
  const MobileCardView = () => (
    <Box>
      {searchable && (
        <Box sx={{ mb: 2 }}>
          <TextField
            fullWidth
            size="small"
            placeholder="Search..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            InputProps={{
              startAdornment: (
                <InputAdornment position="start">
                  <SearchIcon />
                </InputAdornment>
              ),
            }}
          />
        </Box>
      )}
      
      {paginatedRows.map((row) => (
        <MobileCard
          key={row.id}
          onClick={() => onRowClick && onRowClick(row)}
        >
          <CardContent>
            <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
              <Box sx={{ flex: 1 }}>
                <Typography variant="h6" gutterBottom>
                  {primaryColumn.format 
                    ? primaryColumn.format(row[primaryColumn.id])
                    : String(row[primaryColumn.id])
                  }
                </Typography>
                
                {columns
                  .filter((col) => !col.hideOnMobile && col.id !== primaryColumn.id)
                  .map((column) => (
                    <Box key={String(column.id)} sx={{ mb: 0.5 }}>
                      <Typography variant="caption" color="text.secondary">
                        {column.mobileLabel || column.label}:
                      </Typography>
                      <Typography variant="body2">
                        {column.format 
                          ? column.format(row[column.id])
                          : String(row[column.id])
                        }
                      </Typography>
                    </Box>
                  ))}
              </Box>
              
              {(actions || expandable) && (
                <Box>
                  {expandable && (
                    <IconButton
                      size="small"
                      onClick={(e) => {
                        e.stopPropagation();
                        toggleRowExpansion(row.id);
                      }}
                    >
                      {expandedRows.has(row.id) ? <CollapseIcon /> : <ExpandIcon />}
                    </IconButton>
                  )}
                  {actions && (
                    <>
                      <IconButton
                        size="small"
                        onClick={(e) => handleMenuClick(e, row.id)}
                      >
                        <MoreIcon />
                      </IconButton>
                      <Menu
                        anchorEl={anchorEl[row.id]}
                        open={Boolean(anchorEl[row.id])}
                        onClose={() => handleMenuClose(row.id)}
                      >
                        {actions.map((action, index) => (
                          <MenuItem
                            key={index}
                            onClick={() => {
                              action.onClick(row);
                              handleMenuClose(row.id);
                            }}
                          >
                            {action.icon && <Box sx={{ mr: 1 }}>{action.icon}</Box>}
                            {action.label}
                          </MenuItem>
                        ))}
                      </Menu>
                    </>
                  )}
                </Box>
              )}
            </Box>
            
            {expandable && renderExpandedContent && (
              <Collapse in={expandedRows.has(row.id)}>
                <Box sx={{ mt: 2, pt: 2, borderTop: 1, borderColor: 'divider' }}>
                  {renderExpandedContent(row)}
                </Box>
              </Collapse>
            )}
          </CardContent>
        </MobileCard>
      ))}
      
      <TablePagination
        rowsPerPageOptions={[10, 25, 100]}
        component="div"
        count={filteredRows.length}
        rowsPerPage={rowsPerPage}
        page={page}
        onPageChange={handleChangePage}
        onRowsPerPageChange={handleChangeRowsPerPage}
      />
    </Box>
  );

  return isMobile ? <MobileCardView /> : <DesktopTable />;
}
