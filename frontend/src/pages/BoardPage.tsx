import {
    Link as ConnectIcon,
    StickyNote2 as PostItIcon,
    Person as SuspectIcon
} from '@mui/icons-material';
import {
    Alert,
    Box,
    CircularProgress,
    SpeedDial,
    SpeedDialAction,
    SpeedDialIcon,
    TextField,
    Typography
} from '@mui/material';
import { useQuery } from '@tanstack/react-query';
import React, { useCallback, useEffect, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useSocket } from '../contexts/SocketContext';
import { boardsApi } from '../services/api';

const BoardPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { joinBoard, leaveBoard } = useSocket();
  const [selectedItems, setSelectedItems] = useState<Set<string>>(new Set());
  const [isConnecting, setIsConnecting] = useState(false);
  const [items, setItems] = useState<Array<{
    id: string;
    type: 'post-it' | 'suspect-card';
    x: number;
    y: number;
    width: number;
    height: number;
    rotation: number;
    z_index: number;
    content: string;
    color?: string;
    serverId?: string; // when created on server
  }>>([]);
  const [connections, setConnections] = useState<Array<{
    id: string;
    from_item_id: string;
    to_item_id: string;
  }>>([]);

  const canvasRef = useRef<HTMLDivElement | null>(null);
  const [dragging, setDragging] = useState<{
    itemId: string | null;
    offsetX: number;
    offsetY: number;
  }>({ itemId: null, offsetX: 0, offsetY: 0 });
  const [firstConnectId, setFirstConnectId] = useState<string | null>(null);

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
    if (!id) return;
    const now = Date.now();
    const tempId = `temp-${now}`;
    const position = { x: 100 + Math.random() * 200, y: 140 + Math.random() * 120 };
    const newItem = {
      id: tempId,
      type: 'post-it' as const,
      x: position.x,
      y: position.y,
      width: 200,
      height: 200,
      rotation: 0,
      z_index: items.length + 1,
      content: 'Evidence notes...\n',
      color: '#ffeb3b',
    };
    setItems((prev) => [...prev, newItem]);

    // Best-effort server create (backend expects type 'note')
    boardsApi
      .createBoardItem(id, {
        type: 'note' as any,
        x: newItem.x,
        y: newItem.y,
        width: newItem.width,
        height: newItem.height,
        content: newItem.content,
        color: newItem.color,
        metadata: { variant: 'post-it' },
      })
      .then((created) => {
        setItems((prev) => prev.map((it) => (it.id === tempId ? { ...it, serverId: created.id } : it)));
      })
      .catch(() => {
        // Keep local-only on failure
      });
  };

  const handleAddSuspectCard = () => {
    if (!id) return;
    const now = Date.now();
    const tempId = `temp-${now}`;
    const position = { x: 180 + Math.random() * 240, y: 180 + Math.random() * 120 };
    const newItem = {
      id: tempId,
      type: 'suspect-card' as const,
      x: position.x,
      y: position.y,
      width: 250,
      height: 300,
      rotation: 0,
      z_index: items.length + 1,
      content: 'Suspect Name\nAge: Unknown\nLast seen:\nNotes:',
      color: '#f5f5f5',
    };
    setItems((prev) => [...prev, newItem]);

    // Best-effort server create (store as note with metadata)
    boardsApi
      .createBoardItem(id, {
        type: 'note' as any,
        x: newItem.x,
        y: newItem.y,
        width: newItem.width,
        height: newItem.height,
        content: newItem.content,
        color: newItem.color,
        metadata: { variant: 'suspect-card' },
      })
      .then((created) => {
        setItems((prev) => prev.map((it) => (it.id === tempId ? { ...it, serverId: created.id } : it)));
      })
      .catch(() => {
        // Keep local-only on failure
      });
  };

  const handleToggleConnect = () => {
    setIsConnecting(!isConnecting);
    setFirstConnectId(null);
  };

  // Initialize local items/connections when board loads
  useEffect(() => {
    if (!board) return;
    const mappedItems = (board.items || []).map((it: any) => {
      // Determine UI type from metadata.variant (fallback to post-it)
      let meta: any = undefined;
      if (it && typeof it.metadata === 'object') {
        meta = it.metadata;
      } else if (it && typeof it.metadata === 'string') {
        try {
          const decoded = atob(it.metadata);
          meta = JSON.parse(decoded);
        } catch {
          try {
            meta = JSON.parse(it.metadata);
          } catch {}
        }
      }
      const variant = meta?.variant;
      const uiType: 'post-it' | 'suspect-card' = variant === 'suspect-card' ? 'suspect-card' : 'post-it';
      return {
        id: it.id, // Use server ID as frontend ID for loaded items
        type: uiType,
        x: it.x,
        y: it.y,
        width: it.width,
        height: it.height,
        rotation: (it as any).rotation ?? 0,
        z_index: it.z_index ?? 1,
        content: it.content || '',
        color: (it as any).color,
        serverId: it.id,
      };
    });
    setItems(mappedItems);
    
    // Map connections using the same IDs as the items
    const mappedConnections = (board.connections || []).map((c) => ({
      id: c.id,
      from_item_id: c.from_item_id, // These should match the item.id values now
      to_item_id: c.to_item_id,     // These should match the item.id values now
    }));
    setConnections(mappedConnections);
  }, [board]);

  const onItemMouseDown = useCallback((e: React.MouseEvent, itemId: string) => {
    if (isConnecting) return;
    if (!canvasRef.current) return;
    const rect = canvasRef.current.getBoundingClientRect();
    const item = items.find((it) => it.id === itemId);
    if (!item) return;
    const mouseX = e.clientX - rect.left;
    const mouseY = e.clientY - rect.top;
    setDragging({ itemId, offsetX: mouseX - item.x, offsetY: mouseY - item.y });
  }, [isConnecting, items]);

  const onCanvasMouseMove = useCallback((e: React.MouseEvent) => {
    if (!dragging.itemId || !canvasRef.current) return;
    const rect = canvasRef.current.getBoundingClientRect();
    const x = e.clientX - rect.left - dragging.offsetX;
    const y = e.clientY - rect.top - dragging.offsetY;
    setItems((prev) => prev.map((it) => (it.id === dragging.itemId ? { ...it, x, y } : it)));
  }, [dragging]);

  const onCanvasMouseUp = useCallback(() => {
    if (!dragging.itemId) return;
    const moved = items.find((it) => it.id === dragging.itemId);
    setDragging({ itemId: null, offsetX: 0, offsetY: 0 });
    if (!moved || !moved.serverId || !id) return;
    // Persist position best-effort
    boardsApi.updateBoardItem(id, moved.serverId, { x: moved.x, y: moved.y }).catch(() => {});
  }, [dragging.itemId, id, items]);

  const onItemClick = useCallback((itemId: string) => {
    if (!isConnecting) return;
    if (!firstConnectId) {
      setFirstConnectId(itemId);
    } else if (firstConnectId !== itemId) {
      const localId = `temp-conn-${Date.now()}`;
      const newConn = {
        id: localId,
        from_item_id: firstConnectId,
        to_item_id: itemId,
      };
      setConnections((prev) => [...prev, newConn]);
      setFirstConnectId(null);
      // Persist to backend if both ends are server-backed
      if (!id) return;
      const from = items.find((it) => it.id === firstConnectId);
      const to = items.find((it) => it.id === itemId);
      if (!from?.serverId || !to?.serverId) return;
      boardsApi
        .createBoardConnection(id, {
          from_item_id: from.serverId,
          to_item_id: to.serverId,
        } as any)
        .then((created) => {
          // Update connection with server data, keeping frontend item IDs for rendering
          setConnections((prev) => prev.map((c) => (c.id === localId ? { 
            id: created.id, 
            from_item_id: firstConnectId, // Keep frontend IDs for rendering
            to_item_id: itemId 
          } : c)));
        })
        .catch(() => {
          // Revert local if failed
          setConnections((prev) => prev.filter((c) => c.id !== localId));
        });
    }
  }, [firstConnectId, isConnecting, items, id]);

  const handleUpdateContent = useCallback((itemId: string, newContent: string) => {
    setItems((prev) => prev.map((it) => it.id === itemId ? { ...it, content: newContent } : it));
    const item = items.find((it) => it.id === itemId);
    if (!item || !item.serverId || !id) return;
    boardsApi.updateBoardItem(id, item.serverId, { content: newContent }).catch(() => {});
  }, [id, items]);

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
        <Typography variant="h5" component="h1" data-testid="board-title">
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
        ref={canvasRef}
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
        data-testid="board-canvas"
        onMouseMove={onCanvasMouseMove}
        onMouseUp={onCanvasMouseUp}
      >
        {/* Board Items */}
        {items.map((item) => (
          <BoardItemComponent
            key={item.id}
            item={item}
            isSelected={selectedItems.has(item.id)}
            isConnecting={isConnecting}
            canEdit={canEdit}
            onMouseDown={(e) => onItemMouseDown(e, item.id)}
            onClick={() => onItemClick(item.id)}
            onUpdateContent={(content) => handleUpdateContent(item.id, content)}
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
          {connections.map((connection) => {
            const fromItem = items.find(item => item.id === connection.from_item_id);
            const toItem = items.find(item => item.id === connection.to_item_id);
            
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
                data-testid="board-connection"
              />
            );
          })}
        </svg>
      </Box>

      {/* Floating Action Buttons */}
      {canEdit && (
        <SpeedDial
          ariaLabel="Board actions"
          sx={{ position: 'absolute', top: 96, left: 16, zIndex: 1100 }}
          icon={<SpeedDialIcon />}
        >
          <SpeedDialAction
            icon={<PostItIcon />}
            tooltipTitle="Add Post-it Note"
            onClick={handleAddPostIt}
            data-testid="add-post-it"
          />
          <SpeedDialAction
            icon={<SuspectIcon />}
            tooltipTitle="Add Suspect Card"
            onClick={handleAddSuspectCard}
            data-testid="add-suspect-card"
          />
          <SpeedDialAction
            icon={<ConnectIcon />}
            tooltipTitle={isConnecting ? "Exit Connect Mode" : "Connect Items"}
            onClick={handleToggleConnect}
            sx={{ bgcolor: isConnecting ? 'error.main' : 'primary.main' }}
            data-testid="connect-mode-button"
          />
        </SpeedDial>
      )}
    </Box>
  );
};

// Simple board item component for now
interface BoardItemComponentProps {
  item: {
    id: string;
    type: 'post-it' | 'suspect-card';
    x: number;
    y: number;
    width: number;
    height: number;
    rotation: number;
    z_index: number;
    content: string;
    color?: string;
  };
  isSelected: boolean;
  isConnecting: boolean;
  canEdit: boolean;
  onMouseDown: (e: React.MouseEvent) => void;
  onClick: () => void;
  onUpdateContent: (content: string) => void;
}

const BoardItemComponent: React.FC<BoardItemComponentProps> = ({
  item,
  isSelected,
  isConnecting,
  canEdit,
  onMouseDown,
  onClick,
  onUpdateContent,
}) => {
  const isPostIt = item.type === 'post-it';
  const [isEditing, setIsEditing] = React.useState(false);
  const [draft, setDraft] = React.useState(item.content || '');

  React.useEffect(() => {
    setDraft(item.content || '');
  }, [item.content]);

  const commit = React.useCallback(() => {
    setIsEditing(false);
    const trimmed = draft.replace(/\s+$/g, '');
    onUpdateContent(trimmed);
  }, [draft, onUpdateContent]);

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
        cursor: isEditing ? 'text' : (canEdit ? 'move' : 'default'),
        bgcolor: isPostIt ? (item.color || '#ffeb3b') : (item.color || '#f5f5f5'),
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
      onMouseDown={canEdit ? onMouseDown : undefined}
      onClick={onClick}
      data-testid="board-item"
      onDoubleClick={() => canEdit && setIsEditing(true)}
    >
      {/* Pushpin (for all item types) */}
      <Box
        sx={{
          position: 'absolute',
          width: 16,
          height: 16,
          background: 'radial-gradient(circle at 30% 30%, #ff4444, #cc0000)',
          borderRadius: '50% 50% 50% 0',
          transform: 'translateX(-50%) rotate(-45deg)',
          boxShadow: '0 2px 4px rgba(0,0,0,0.3), inset 1px 1px 2px rgba(255,255,255,0.3)',
          top: -8,
          left: '50%',
          zIndex: 10,
        }}
      />

      {/* Content */}
      {isEditing ? (
        <TextField
          fullWidth
          multiline
          size="small"
          variant="outlined"
          value={draft}
          onChange={(e) => setDraft(e.target.value)}
          onBlur={commit}
          onKeyDown={(e) => {
            if (e.key === 'Enter' && (e.shiftKey || isPostIt)) {
              // Allow Shift+Enter for newline, plain Enter commits for post-it
              if (!e.shiftKey) {
                e.preventDefault();
                commit();
              }
            } else if (e.key === 'Enter' && !isPostIt) {
              // For suspect card, Enter commits as well
              e.preventDefault();
              commit();
            } else if (e.key === 'Escape') {
              e.preventDefault();
              setIsEditing(false);
              setDraft(item.content || '');
            }
          }}
          autoFocus
          inputProps={{ 'data-testid': 'item-edit-input' }}
          sx={{
            '& .MuiInputBase-input': {
              fontFamily: isPostIt ? '"Courier New", monospace' : undefined,
              fontSize: isPostIt ? 14 : undefined,
              lineHeight: isPostIt ? 1.4 : undefined,
              color: '#333',
              whiteSpace: 'pre-wrap',
              padding: isPostIt ? 0 : undefined,
            },
            backgroundColor: 'rgba(255,255,255,0.9)'
          }}
        />
      ) : (
        isPostIt ? (
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
              {(item.content?.split('\n')[0] || 'Suspect Name')}
            </Typography>
            <Typography variant="body2" color="textSecondary" sx={{ whiteSpace: 'pre-wrap' }}>
              {item.content || 'Age: Unknown\nLast seen:\nNotes:'}
            </Typography>
          </Box>
        )
      )}
    </Box>
  );
};

export default BoardPage;


