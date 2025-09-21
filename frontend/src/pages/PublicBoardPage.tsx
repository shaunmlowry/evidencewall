import {
    Alert,
    AppBar,
    Box,
    CircularProgress,
    Container,
    Toolbar,
    Typography,
} from '@mui/material';
import { useQuery } from '@tanstack/react-query';
import React from 'react';
import { useParams } from 'react-router-dom';
import { boardsApi } from '../services/api';

const PublicBoardPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();

  // Fetch public board data
  const { data: board, isLoading, error } = useQuery({
    queryKey: ['publicBoard', id],
    queryFn: () => boardsApi.getPublicBoard(id!),
    enabled: !!id,
  });

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="100vh">
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Container>
        <Box p={3}>
          <Alert severity="error">
            Failed to load board. This board may not be public or may not exist.
          </Alert>
        </Box>
      </Container>
    );
  }

  if (!board) {
    return (
      <Container>
        <Box p={3}>
          <Alert severity="warning">Board not found.</Alert>
        </Box>
      </Container>
    );
  }

  return (
    <Box sx={{ height: '100vh', overflow: 'hidden' }}>
      {/* Header */}
      <AppBar position="static">
        <Toolbar>
          <Typography variant="h6" component="div" sx={{ flexGrow: 1 }}>
            Evidence Wall - {board.title}
          </Typography>
          <Typography variant="body2">
            Public Board (Read Only)
          </Typography>
        </Toolbar>
      </AppBar>

      {/* Board Description */}
      {board.description && (
        <Box sx={{ p: 2, bgcolor: 'background.paper', borderBottom: 1, borderColor: 'divider' }}>
          <Typography variant="body2" color="textSecondary">
            {board.description}
          </Typography>
        </Box>
      )}

      {/* Board Canvas */}
      <Box
        sx={{
          position: 'relative',
          height: 'calc(100vh - 64px)', // Account for AppBar
          background: `
            radial-gradient(circle at 20% 30%, rgba(139, 119, 101, 0.8), transparent 2px),
            radial-gradient(circle at 70% 20%, rgba(160, 142, 124, 0.6), transparent 2px),
            radial-gradient(circle at 40% 70%, rgba(120, 100, 85, 0.7), transparent 2px),
            radial-gradient(circle at 90% 80%, rgba(145, 125, 108, 0.5), transparent 2px),
            linear-gradient(45deg, #c4a484 0%, #d4b896 25%, #c8a688 50%, #ddc2a4 75%, #c4a484 100%)
          `,
          backgroundSize: '50px 50px, 75px 75px, 60px 60px, 40px 40px, 100% 100%',
          overflow: 'auto',
        }}
      >
        {/* Board Items */}
        {board.items?.map((item: any) => {
          let meta: any = undefined;
          if (item && typeof item.metadata === 'object') {
            meta = item.metadata;
          } else if (item && typeof item.metadata === 'string') {
            try {
              const decoded = atob(item.metadata);
              meta = JSON.parse(decoded);
            } catch {
              try { meta = JSON.parse(item.metadata); } catch {}
            }
          }
          const variant = meta?.variant;
          const isPostIt = variant !== 'suspect-card';
          
          return (
            <Box
              key={item.id}
              sx={{
                position: 'absolute',
                left: item.x,
                top: item.y,
                width: item.width,
                height: item.height,
                transform: `rotate(${item.rotation}deg)`,
                zIndex: item.z_index,
                bgcolor: isPostIt ? (item.color || '#ffeb3b') : (item.color || '#f5f5f5'),
                border: isPostIt ? 'none' : '1px solid #ddd',
                borderRadius: isPostIt ? '2px' : '4px',
                boxShadow: isPostIt 
                  ? '0 4px 8px rgba(0,0,0,0.2), inset 0 0 0 1px rgba(0,0,0,0.1)'
                  : '0 6px 12px rgba(0,0,0,0.3), inset 0 0 0 1px rgba(0,0,0,0.1)',
                p: isPostIt ? '20px 15px 15px 15px' : 2,
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
        })}

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

        {/* Empty state */}
        {(!board.items || board.items.length === 0) && (
          <Box
            display="flex"
            justifyContent="center"
            alignItems="center"
            height="100%"
          >
            <Typography variant="h6" color="textSecondary">
              This board is empty
            </Typography>
          </Box>
        )}
      </Box>
    </Box>
  );
};

export default PublicBoardPage;


