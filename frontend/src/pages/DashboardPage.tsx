import {
    Add as AddIcon,
    Lock as LockIcon,
    MoreVert as MoreVertIcon,
    People as PeopleIcon,
    Public as PublicIcon,
} from '@mui/icons-material';
import {
    Alert,
    Box,
    Button,
    Card,
    CardActions,
    CardContent,
    Chip,
    Container,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    FormControl,
    Grid,
    IconButton,
    InputLabel,
    Menu,
    MenuItem,
    Select,
    TextField,
    Typography,
} from '@mui/material';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import { boardsApi } from '../services/api';
import { Board, BoardVisibility, CreateBoardRequest } from '../types';

const DashboardPage: React.FC = () => {
  const navigate = useNavigate();
  const { user } = useAuth();
  const queryClient = useQueryClient();
  
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [selectedBoard, setSelectedBoard] = useState<Board | null>(null);
  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [newBoard, setNewBoard] = useState<CreateBoardRequest>({
    title: '',
    description: '',
    visibility: 'private',
  });
  const [error, setError] = useState('');

  // Fetch boards
  const { data: boardsResponse, isLoading } = useQuery({
    queryKey: ['boards'],
    queryFn: () => boardsApi.getBoards(),
  });

  // Create board mutation
  const createBoardMutation = useMutation({
    mutationFn: boardsApi.createBoard,
    onSuccess: (board) => {
      queryClient.invalidateQueries({ queryKey: ['boards'] });
      setCreateDialogOpen(false);
      setNewBoard({ title: '', description: '', visibility: 'private' });
      navigate(`/board/${board.id}`);
    },
    onError: (err: any) => {
      setError(err.response?.data?.error || 'Failed to create board');
    },
  });

  // Delete board mutation
  const deleteBoardMutation = useMutation({
    mutationFn: boardsApi.deleteBoard,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['boards'] });
      setAnchorEl(null);
      setSelectedBoard(null);
    },
  });

  const handleMenuOpen = (event: React.MouseEvent<HTMLElement>, board: Board) => {
    setAnchorEl(event.currentTarget);
    setSelectedBoard(board);
  };

  const handleMenuClose = () => {
    setAnchorEl(null);
    setSelectedBoard(null);
  };

  const handleCreateBoard = () => {
    if (!newBoard.title.trim()) {
      setError('Board title is required');
      return;
    }
    createBoardMutation.mutate(newBoard);
  };

  const handleDeleteBoard = () => {
    if (selectedBoard) {
      deleteBoardMutation.mutate(selectedBoard.id);
    }
  };

  const getVisibilityIcon = (visibility: BoardVisibility) => {
    switch (visibility) {
      case 'public':
        return <PublicIcon fontSize="small" />;
      case 'shared':
        return <PeopleIcon fontSize="small" />;
      default:
        return <LockIcon fontSize="small" />;
    }
  };

  const getVisibilityColor = (visibility: BoardVisibility) => {
    switch (visibility) {
      case 'public':
        return 'success';
      case 'shared':
        return 'info';
      default:
        return 'default';
    }
  };

  const canDeleteBoard = (board: Board) => {
    return board.owner_id === user?.id || board.permission === 'admin';
  };

  if (isLoading) {
    return (
      <Container>
        <Box display="flex" justifyContent="center" alignItems="center" minHeight="50vh">
          <Typography>Loading boards...</Typography>
        </Box>
      </Container>
    );
  }

  const boards = boardsResponse?.data || [];

  return (
    <Container maxWidth="lg">
      <Box mb={4}>
        <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
          <Typography variant="h4" component="h1" data-testid="dashboard-title">
            My Boards
          </Typography>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => setCreateDialogOpen(true)}
            data-testid="create-board-button"
          >
            Create Board
          </Button>
        </Box>

        {boards.length === 0 ? (
          <Box textAlign="center" py={8}>
            <Typography variant="h6" color="textSecondary" gutterBottom>
              No boards yet
            </Typography>
            <Typography variant="body2" color="textSecondary" mb={3}>
              Create your first evidence board to get started
            </Typography>
            <Button
              variant="contained"
              startIcon={<AddIcon />}
              onClick={() => setCreateDialogOpen(true)}
            >
              Create Your First Board
            </Button>
          </Box>
        ) : (
          <Grid container spacing={3}>
            {boards.map((board) => (
              <Grid item xs={12} sm={6} md={4} key={board.id}>
                <Card>
                  <CardContent>
                    <Box display="flex" justifyContent="space-between" alignItems="flex-start" mb={2}>
                      <Typography variant="h6" component="h2" noWrap>
                        {board.title}
                      </Typography>
                      <IconButton
                        size="small"
                        onClick={(e) => handleMenuOpen(e, board)}
                      >
                        <MoreVertIcon />
                      </IconButton>
                    </Box>
                    
                    {board.description && (
                      <Typography variant="body2" color="textSecondary" mb={2} noWrap>
                        {board.description}
                      </Typography>
                    )}

                    <Box display="flex" gap={1} mb={2}>
                      <Chip
                        icon={getVisibilityIcon(board.visibility)}
                        label={board.visibility}
                        size="small"
                        color={getVisibilityColor(board.visibility) as any}
                      />
                      {board.permission && (
                        <Chip
                          label={board.permission.replace('_', ' ')}
                          size="small"
                          variant="outlined"
                        />
                      )}
                    </Box>

                    <Typography variant="caption" color="textSecondary">
                      Updated {new Date(board.updated_at).toLocaleDateString()}
                    </Typography>
                  </CardContent>
                  <CardActions>
                    <Button
                      size="small"
                      onClick={() => navigate(`/board/${board.id}`)}
                    >
                      Open
                    </Button>
                  </CardActions>
                </Card>
              </Grid>
            ))}
          </Grid>
        )}
      </Box>

      {/* Board menu */}
      <Menu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={handleMenuClose}
      >
        <MenuItem onClick={() => navigate(`/board/${selectedBoard?.id}`)}>
          Open
        </MenuItem>
        {selectedBoard && canDeleteBoard(selectedBoard) && (
          <MenuItem onClick={handleDeleteBoard} sx={{ color: 'error.main' }}>
            Delete
          </MenuItem>
        )}
      </Menu>

      {/* Create board dialog */}
      <Dialog open={createDialogOpen} onClose={() => setCreateDialogOpen(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Create New Board</DialogTitle>
        <DialogContent>
          {error && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {error}
            </Alert>
          )}
          <TextField
            autoFocus
            margin="dense"
            label="Board Title"
            fullWidth
            variant="outlined"
            value={newBoard.title}
            onChange={(e) => setNewBoard({ ...newBoard, title: e.target.value })}
            sx={{ mb: 2 }}
            inputProps={{ 'data-testid': 'board-title-input' }}
          />
          <TextField
            margin="dense"
            label="Description (optional)"
            fullWidth
            multiline
            rows={3}
            variant="outlined"
            value={newBoard.description}
            onChange={(e) => setNewBoard({ ...newBoard, description: e.target.value })}
            sx={{ mb: 2 }}
            inputProps={{ 'data-testid': 'board-description-input' }}
          />
          <FormControl fullWidth>
            <InputLabel>Visibility</InputLabel>
            <Select
              value={newBoard.visibility}
              label="Visibility"
              onChange={(e) => setNewBoard({ ...newBoard, visibility: e.target.value as BoardVisibility })}
            >
              <MenuItem value="private">Private - Only you can access</MenuItem>
              <MenuItem value="shared">Shared - Invite specific users</MenuItem>
              <MenuItem value="public">Public - Anyone can view</MenuItem>
            </Select>
          </FormControl>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setCreateDialogOpen(false)}>Cancel</Button>
          <Button
            onClick={handleCreateBoard}
            variant="contained"
            disabled={createBoardMutation.isPending}
            data-testid="create-board-submit"
          >
            {createBoardMutation.isPending ? 'Creating...' : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>
    </Container>
  );
};

export default DashboardPage;


