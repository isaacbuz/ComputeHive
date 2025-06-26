import React, { useState, useMemo } from 'react';
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
import { useResponsive, useResponsiveStyles } from '../../hooks/useResponsive';
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
  sortable?: boolean;
  width?: string;
  render?: (value: any, row: any) => React.ReactNode;
  mobile?: {
    show?: boolean;
    label?: string;
    render?: (value: any, row: any) => React.ReactNode;
  };
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
  className?: string;
  sortable?: boolean;
  pagination?: {
    enabled: boolean;
    pageSize?: number;
    currentPage?: number;
    onPageChange?: (page: number) => void;
  };
  loading?: boolean;
  emptyMessage?: string;
  selectable?: boolean;
  onSelectionChange?: (selectedRows: T[]) => void;
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
  className = '',
  sortable = true,
  pagination,
  loading = false,
  emptyMessage = 'No data available',
  selectable = false,
  onSelectionChange,
}: ResponsiveDataTableProps<T>) {
  const { isMobile } = useResponsive();
  const theme = useTheme();
  const { dataDisplay } = useResponsiveStyles();
  const [page, setPage] = useState(0);
  const [rowsPerPage, setRowsPerPage] = useState(10);
  const [searchQuery, setSearchQuery] = useState('');
  const [expandedRows, setExpandedRows] = useState<Set<string | number>>(new Set());
  const [anchorEl, setAnchorEl] = useState<{ [key: string]: HTMLElement | null }>({});
  const [sortConfig, setSortConfig] = useState<{
    key: keyof T;
    direction: 'asc' | 'desc';
  } | null>(null);
  const [selectedRows, setSelectedRows] = useState<T[]>([]);

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

  // Filter and sort data
  const processedData = useMemo(() => {
    let filteredData = rows;

    // Apply search filter
    if (searchQuery) {
      filteredData = rows.filter((row) =>
        columns.some((column) =>
          String(row[column.id]).toLowerCase().includes(searchQuery.toLowerCase())
        )
      );
    }

    // Apply sorting
    if (sortConfig && sortable) {
      filteredData = [...filteredData].sort((a, b) => {
        const aValue = a[sortConfig.key];
        const bValue = b[sortConfig.key];

        if (aValue < bValue) {
          return sortConfig.direction === 'asc' ? -1 : 1;
        }
        if (aValue > bValue) {
          return sortConfig.direction === 'asc' ? 1 : -1;
        }
        return 0;
      });
    }

    return filteredData;
  }, [rows, sortConfig, searchQuery, sortable]);

  // Handle sorting
  const handleSort = (key: keyof T) => {
    if (!sortable) return;

    setSortConfig((current) => {
      if (current?.key === key) {
        return {
          key,
          direction: current.direction === 'asc' ? 'desc' : 'asc',
        };
      }
      return { key, direction: 'asc' };
    });
  };

  // Handle row selection
  const handleRowSelect = (row: T) => {
    if (!selectable) return;

    const newSelectedRows = selectedRows.includes(row)
      ? selectedRows.filter((r) => r !== row)
      : [...selectedRows, row];

    setSelectedRows(newSelectedRows);
    onSelectionChange?.(newSelectedRows);
  };

  // Handle select all
  const handleSelectAll = () => {
    if (!selectable) return;

    const newSelectedRows = selectedRows.length === processedData.length ? [] : processedData;
    setSelectedRows(newSelectedRows);
    onSelectionChange?.(newSelectedRows);
  };

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
              {selectable && (
                <TableCell className="table-header select-header">
                  <input
                    type="checkbox"
                    checked={selectedRows.length === processedData.length && processedData.length > 0}
                    onChange={handleSelectAll}
                  />
                </TableCell>
              )}
              {columns.map((column) => (
                <TableCell
                  key={String(column.id)}
                  align={column.align}
                  style={{ minWidth: column.minWidth }}
                  className={`table-header ${column.sortable ? 'sortable' : ''}`}
                  onClick={() => column.sortable && handleSort(column.id as keyof T)}
                >
                  <div className="header-content">
                    {column.label}
                    {column.sortable && sortConfig?.key === column.id && (
                      <span className="sort-indicator">
                        {sortConfig.direction === 'asc' ? '↑' : '↓'}
                      </span>
                    )}
                  </div>
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
                  {selectable && (
                    <TableCell className="table-cell select-cell">
                      <input
                        type="checkbox"
                        checked={selectedRows.includes(row)}
                        onChange={() => handleRowSelect(row)}
                        onClick={(e) => e.stopPropagation()}
                      />
                    </TableCell>
                  )}
                  {columns.map((column) => {
                    const value = row[column.id];
                    return (
                      <TableCell key={String(column.id)} align={column.align}>
                        {column.render ? column.render(value, row) : String(value)}
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
                        {column.mobile?.render
                          ? column.mobile.render(row[column.id], row)
                          : column.render
                          ? column.render(row[column.id], row)
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

  // Search component
  const SearchComponent = () => {
    if (!searchable) return null;

    return (
      <div className="search-container">
        <input
          type="text"
          placeholder="Search..."
          value={searchQuery}
          onChange={(e) => {
            setSearchQuery(e.target.value);
          }}
          className="search-input"
        />
      </div>
    );
  };

  // Pagination component
  const PaginationComponent = () => {
    if (!pagination?.enabled) return null;

    const totalPages = Math.ceil(processedData.length / (pagination.pageSize || 10));
    const currentPage = pagination.currentPage || 1;

    return (
      <div className="pagination-container">
        <button
          className="pagination-button"
          disabled={currentPage === 1}
          onClick={() => pagination.onPageChange?.(currentPage - 1)}
        >
          Previous
        </button>
        
        <span className="pagination-info">
          Page {currentPage} of {totalPages}
        </span>
        
        <button
          className="pagination-button"
          disabled={currentPage === totalPages}
          onClick={() => pagination.onPageChange?.(currentPage + 1)}
        >
          Next
        </button>
      </div>
    );
  };

  if (loading) {
    return (
      <div className={`responsive-data-table loading ${className}`}>
        <div className="loading-spinner">Loading...</div>
      </div>
    );
  }

  return (
    <div className={`responsive-data-table ${className}`}>
      <SearchComponent />
      
      {processedData.length === 0 ? (
        <div className="empty-state">
          <p className="empty-message">{emptyMessage}</p>
        </div>
      ) : (
        <>
          {dataDisplay.tableType === 'table' ? <DesktopTable /> : <MobileCardView />}
          <PaginationComponent />
        </>
      )}
    </div>
  );
}

export default ResponsiveDataTable; 