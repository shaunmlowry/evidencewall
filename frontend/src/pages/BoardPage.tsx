import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import {
  Box,
  Typography,
  CircularProgress,
  Alert,
  Fab,
  SpeedDial,
  SpeedDialAction,
  SpeedDialIcon,
} from '@mui/material';
import {
  Add as AddIcon,
  StickyNote2 as PostItIcon,
  Person as SuspectIcon,
  Link as ConnectIcon,
} from '@mui/icons-material';
import { useQuery } from '@tanstack/react-query';
import { boardsApi } from '../services/api';
import { useSocket } from '../contexts/SocketContext';
import { Board, BoardItem, BoardConnection } from '../types';

const BoardPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { joinBoard, leaveBoard } = useSocket();
  const [selectedItems, setSelectedItems] = useState<Set<string>>(new Set());
  const [isConnecting, setIsConnecting] = useState(false);

  // Fetch board data
  const { data: board, isLoading, error } = useQuery({
    queryKey: ['board', id],
    queryFn: () => boardsApi.getBoard(id!),
    enabled: !!id,
  });

  // Join board room on mount
  useEffect(() => {
    if (id) {
      joinBoard(id);
      return () => leaveBoard(id);
    }
  }, [id, joinBoard, leaveBoard]);

  const handleAddPostIt = () => {
    // TODO: Implement add post-it functionality
    console.log('Add post-it');
  };

  const handleAddSuspectCard = () => {
    // TODO: Implement add suspect card functionality
    console.log('Add suspect card');
  };

  const handleToggleConnect = () => {
    setIsConnecting(!isConnecting);
  };

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="50vh">
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Box p={3}>
        <Alert severity="error">
          Failed to load board. You may not have permission to view this board.
        </Alert>
      </Box>
    );
  }

  if (!board) {
    return (
      <Box p={3}>
        <Alert severity="warning">Board not found.</Alert>
      </Box>
    );
  }

  const canEdit = board.permission === 'read_write' || board.permission === 'admin';

  return (
    <Box sx={{ height: '100vh', overflow: 'hidden', position: 'relative' }}>
      {/* Board Header */}
      <Box
        sx={{
          position: 'absolute',
          top: 0,
          left: 0,
          right: 0,
          zIndex: 1000,
          bgcolor: 'background.paper',
          borderBottom: 1,
          borderColor: 'divider',
          p: 2,
        }}
      >
        <Typography variant="h5" component="h1">
          {board.title}
        </Typography>
        {board.description && (
          <Typography variant="body2" color="textSecondary">
            {board.description}
          </Typography>
        )}
      </Box>

      {/* Board Canvas */}
      <Box
        sx={{
          position: 'absolute',
          top: 80, // Account for header
          left: 0,
          right: 0,
          bottom: 0,
          background: `
            radial-gradient(circle at 20% 30%, rgba(139, 119, 101, 0.8), transparent 2px),
            radial-gradient(circle at 70% 20%, rgba(160, 142, 124, 0.6), transparent 2px),
            radial-gradient(circle at 40% 70%, rgba(120, 100, 85, 0.7), transparent 2px),
            radial-gradient(circle at 90% 80%, rgba(145, 125, 108, 0.5), transparent 2px),
            linear-gradient(45deg, #c4a484 0%, #d4b896 25%, #c8a688 50%, #ddc2a4 75%, #c4a484 100%)
          `,
          backgroundSize: '50px 50px, 75px 75px, 60px 60px, 40px 40px, 100% 100%',
          cursor: isConnecting ? 'crosshair' : 'default',
          overflow: 'hidden',
        }}
      >
        {/* Board Items */}
        {board.items?.map((item) => (
          <BoardItemComponent
            key={item.id}
            item={item}
            isSelected={selectedItems.has(item.id)}
            isConnecting={isConnecting}
            canEdit={canEdit}
          />
        ))}

        {/* Board Connections */}
        <svg
          style={{
            position: 'absolute',
            top: 0,
            left: 0,
            width: '100%',
            height: '100%',
            pointerEvents: 'none',
            zIndex: 50,
          }}
        >
          {board.connections?.map((connection) => {
            const fromItem = board.items?.find(item => item.id === connection.from_item_id);
            const toItem = board.items?.find(item => item.id === connection.to_item_id);
            
            if (!fromItem || !toItem) return null;

            const x1 = fromItem.x + fromItem.width / 2;
            const y1 = fromItem.y;
            const x2 = toItem.x + toItem.width / 2;
            const y2 = toItem.y;

            // Calculate sag for realistic string physics
            const distance = Math.sqrt((x2 - x1) ** 2 + (y2 - y1) ** 2);
            const sagAmount = Math.min(distance * 0.1, 30);
            
            const midX = (x1 + x2) / 2;
            const midY = (y1 + y2) / 2 + sagAmount;

            return (
              <path
                key={connection.id}
                d={`M ${x1} ${y1} Q ${midX} ${midY} ${x2} ${y2}`}
                stroke="#cc0000"
                strokeWidth="2"
                fill="none"
                style={{ filter: 'drop-shadow(1px 1px 2px rgba(0,0,0,0.3))' }}
              />
            );
          })}
        </svg>
      </Box>

      {/* Floating Action Buttons */}
      {canEdit && (
        <SpeedDial
          ariaLabel="Board actions"
          sx={{ position: 'absolute', bottom: 16, right: 16 }}
          icon={<SpeedDialIcon />}
        >
          <SpeedDialAction
            icon={<PostItIcon />}
            tooltipTitle="Add Post-it Note"
            onClick={handleAddPostIt}
          />
          <SpeedDialAction
            icon={<SuspectIcon />}
            tooltipTitle="Add Suspect Card"
            onClick={handleAddSuspectCard}
          />
          <SpeedDialAction
            icon={<ConnectIcon />}
            tooltipTitle={isConnecting ? "Exit Connect Mode" : "Connect Items"}
            onClick={handleToggleConnect}
            sx={{ bgcolor: isConnecting ? 'error.main' : 'primary.main' }}
          />
        </SpeedDial>
      )}
    </Box>
  );
};

// Simple board item component for now
interface BoardItemComponentProps {
  item: BoardItem;
  isSelected: boolean;
  isConnecting: boolean;
  canEdit: boolean;
}

const BoardItemComponent: React.FC<BoardItemComponentProps> = ({
  item,
  isSelected,
  isConnecting,
  canEdit,
}) => {
  const isPostIt = item.type === 'post-it';

  return (
    <Box
      sx={{
        position: 'absolute',
        left: item.x,
        top: item.y,
        width: item.width,
        height: item.height,
        transform: `rotate(${item.rotation}deg)`,
        zIndex: item.z_index,
        cursor: canEdit ? 'move' : 'default',
        bgcolor: isPostIt ? '#ffeb3b' : '#f5f5f5',
        border: isPostIt ? 'none' : '1px solid #ddd',
        borderRadius: isPostIt ? '2px' : '4px',
        boxShadow: isPostIt 
          ? '0 4px 8px rgba(0,0,0,0.2), inset 0 0 0 1px rgba(0,0,0,0.1)'
          : '0 6px 12px rgba(0,0,0,0.3), inset 0 0 0 1px rgba(0,0,0,0.1)',
        p: isPostIt ? '20px 15px 15px 15px' : 2,
        outline: isSelected ? '2px solid #1976d2' : 'none',
        '&:hover': canEdit ? {
          transform: `rotate(${item.rotation}deg) scale(1.02)`,
          zIndex: 100,
        } : {},
      }}
    >
      {/* Pin for post-it */}
      {isPostIt && (
        <Box
          sx={{
            position: 'absolute',
            width: 16,
            height: 16,
            background: 'radial-gradient(circle at 30% 30%, #ff4444, #cc0000)',
            borderRadius: '50% 50% 50% 0',
            transform: 'rotate(-45deg)',
            boxShadow: '0 2px 4px rgba(0,0,0,0.3), inset 1px 1px 2px rgba(255,255,255,0.3)',
            top: -8,
            left: '50%',
            marginLeft: -8,
            zIndex: 10,
          }}
        />
      )}

      {/* Content */}
      {isPostIt ? (
        <Typography
          variant="body2"
          sx={{
            fontFamily: '"Courier New", monospace',
            fontSize: 14,
            lineHeight: 1.4,
            color: '#333',
            whiteSpace: 'pre-wrap',
            overflow: 'hidden',
          }}
        >
          {item.content || 'Evidence notes...'}
        </Typography>
      ) : (
        <Box>
          {/* Suspect card photo area */}
          <Box
            sx={{
              width: '100%',
              height: 120,
              bgcolor: '#e0e0e0',
              borderBottom: '1px solid #ddd',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: 48,
              color: '#999',
              mb: 1,
            }}
          >
            👤
          </Box>
          
          {/* Suspect info */}
          <Typography variant="h6" sx={{ fontWeight: 'bold', mb: 1 }}>
            Suspect Name
          </Typography>
          <Typography variant="body2" color="textSecondary">
            {item.content || 'Age: Unknown\nLast seen:\nNotes:'}
          </Typography>
        </Box>
      )}
    </Box>
  );
};

export default BoardPage;


