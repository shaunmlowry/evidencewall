import {
    Link as ConnectIcon,
    StickyNote2 as PostItIcon,
    Person as SuspectIcon,
    ZoomIn as ZoomInIcon,
    ZoomOut as ZoomOutIcon,
    CenterFocusStrong as ResetZoomIcon,
    Delete as DeleteIcon
} from '@mui/icons-material';
import {
    Alert,
    Box,
    CircularProgress,
    SpeedDial,
    SpeedDialAction,
    SpeedDialIcon,
    TextField,
    Typography,
    IconButton,
    Paper
} from '@mui/material';
import { useQuery } from '@tanstack/react-query';
import React, { useCallback, useEffect, useRef, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useSocket } from '../contexts/SocketContext';
import { boardsApi } from '../services/api';

const BoardPage: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { joinBoard, leaveBoard, socket } = useSocket();
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

  // Zoom and scroll state
  const [zoom, setZoom] = useState(1.0);
  const [panX, setPanX] = useState(0);
  const [panY, setPanY] = useState(0);
  const [isPanning, setIsPanning] = useState(false);
  const [panStart, setPanStart] = useState({ x: 0, y: 0, panX: 0, panY: 0 });

  const canvasRef = useRef<HTMLDivElement | null>(null);
  const [dragging, setDragging] = useState<{
    itemId: string | null;
    offsetX: number;
    offsetY: number;
  }>({ itemId: null, offsetX: 0, offsetY: 0 });
  const [firstConnectId, setFirstConnectId] = useState<string | null>(null);

  // Zoom and pan functions
  const zoomIn = useCallback(() => {
    setZoom(prev => Math.min(prev + 0.01, 3.0)); // Max 3x zoom
  }, []);

  const zoomOut = useCallback(() => {
    setZoom(prev => Math.max(prev - 0.01, 0.1)); // Min 0.1x zoom
  }, []);

  const resetZoom = useCallback(() => {
    setZoom(1.0);
    setPanX(0);
    setPanY(0);
  }, []);

  // Transform coordinates from screen to board space
  const screenToBoard = useCallback((screenX: number, screenY: number) => {
    return {
      x: (screenX - panX) / zoom,
      y: (screenY - panY) / zoom
    };
  }, [zoom, panX, panY]);

  // Transform coordinates from board to screen space
  const boardToScreen = useCallback((boardX: number, boardY: number) => {
    return {
      x: boardX * zoom + panX,
      y: boardY * zoom + panY
    };
  }, [zoom, panX, panY]);

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

  // Listen for realtime updates
  useEffect(() => {
    if (!socket || !id) return;
    const handler = (raw: any) => {
      try {
        const msg = typeof raw === 'string' ? JSON.parse(raw) : raw;
        if (!msg || msg.board_id !== id) return;
        switch (msg.event) {
          case 'item_created': {
            const it = msg.data;
            setItems((prev) => {
              if (prev.some((p) => p.id === it.id)) return prev;
              return [
                ...prev,
                {
                  id: it.id,
                  type: ((it.metadata && (JSON.parse(typeof it.metadata === 'string' ? it.metadata : JSON.stringify(it.metadata)).variant)) === 'suspect-card') ? 'suspect-card' : 'post-it',
                  x: it.x,
                  y: it.y,
                  width: it.width,
                  height: it.height,
                  rotation: it.rotation ?? 0,
                  z_index: it.z_index ?? 1,
                  content: it.content || '',
                  color: it.color,
                  serverId: it.id,
                },
              ];
            });
            break;
          }
          case 'item_updated': {
            const it = msg.data;
            setItems((prev) => prev.map((p) => (p.id === it.id ? {
              ...p,
              x: it.x ?? p.x,
              y: it.y ?? p.y,
              width: it.width ?? p.width,
              height: it.height ?? p.height,
              z_index: it.z_index ?? p.z_index,
              content: it.content ?? p.content,
              color: it.color ?? p.color,
            } : p)));
            break;
          }
          case 'item_deleted': {
            const delId = msg.data?.id;
            setItems((prev) => prev.filter((p) => p.id !== delId));
            break;
          }
          case 'connection_created': {
            const c = msg.data;
            setConnections((prev) => prev.some((pc) => pc.id === c.id) ? prev : [...prev, {
              id: c.id,
              from_item_id: c.from_item_id,
              to_item_id: c.to_item_id,
            }]);
            break;
          }
          case 'connection_updated': {
            // Currently style-only, ignore for now
            break;
          }
          case 'connection_deleted': {
            const delId = msg.data?.id;
            setConnections((prev) => prev.filter((c) => c.id !== delId));
            break;
          }
        }
      } catch {}
    };
    socket.on('board_update', handler);
    return () => {
      socket.off('board_update', handler);
    };
  }, [socket, id]);

  // (removed debug global click listener)

  // Utility function to calculate visible viewport bounds in board coordinates
  const getVisibleBounds = () => {
    if (!canvasRef.current) {
      return { minX: 0, minY: 0, maxX: 400, maxY: 300 };
    }
    
    const canvasRect = canvasRef.current.getBoundingClientRect();
    const viewportWidth = canvasRect.width;
    const viewportHeight = canvasRect.height;
    
    // Convert viewport bounds to board coordinates
    // Account for zoom and pan transformations
    const minX = -panX / zoom;
    const minY = -panY / zoom;
    const maxX = minX + viewportWidth / zoom;
    const maxY = minY + viewportHeight / zoom;
    
    return { minX, minY, maxX, maxY };
  };

  // Utility function to get a random position within visible bounds
  const getRandomPositionInViewport = (itemWidth: number, itemHeight: number) => {
    const bounds = getVisibleBounds();
    
    // Ensure the item fits within the visible area with some padding
    const padding = 20;
    const maxX = Math.max(bounds.minX + padding, bounds.maxX - itemWidth - padding);
    const maxY = Math.max(bounds.minY + padding, bounds.maxY - itemHeight - padding);
    
    // If the visible area is too small, place at center of viewport
    if (maxX <= bounds.minX + padding || maxY <= bounds.minY + padding) {
      return {
        x: bounds.minX + (bounds.maxX - bounds.minX) / 2 - itemWidth / 2,
        y: bounds.minY + (bounds.maxY - bounds.minY) / 2 - itemHeight / 2
      };
    }
    
    return {
      x: bounds.minX + padding + Math.random() * (maxX - bounds.minX - padding),
      y: bounds.minY + padding + Math.random() * (maxY - bounds.minY - padding)
    };
  };

  const handleAddPostIt = () => {
    if (!id) return;
    const now = Date.now();
    const tempId = `temp-${now}`;
    // Place items in visible viewport area
    const position = getRandomPositionInViewport(200, 200);
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
    // Place items in visible viewport area
    const position = getRandomPositionInViewport(280, 400);
    const newItem = {
      id: tempId,
      type: 'suspect-card' as const,
      x: position.x,
      y: position.y,
      width: 280,
      height: 400,
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

  const handleDeleteSelected = useCallback(() => {
    if (selectedItems.size === 0) return;
    
    const itemsToDelete = Array.from(selectedItems);
    
    // Remove items locally
    setItems(prev => prev.filter(item => !selectedItems.has(item.id)));
    
    // Remove connections involving deleted items
    setConnections(prev => prev.filter(conn => 
      !itemsToDelete.includes(conn.from_item_id) && 
      !itemsToDelete.includes(conn.to_item_id)
    ));
    
    // Clear selection
    setSelectedItems(new Set());
    
    // Delete from server (best effort)
    if (id) {
      itemsToDelete.forEach(itemId => {
        const item = items.find(it => it.id === itemId);
        if (item?.serverId) {
          boardsApi.deleteBoardItem(id, item.serverId).catch(() => {
            // Ignore deletion errors for now
          });
        }
      });
    }
  }, [selectedItems, items, id]);

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
    if (isConnecting) {
      return;
    }
    if (!canvasRef.current) return;
    const rect = canvasRef.current.getBoundingClientRect();
    const item = items.find((it) => it.id === itemId);
    if (!item) return;
    const mouseX = e.clientX - rect.left;
    const mouseY = e.clientY - rect.top;
    const boardCoords = screenToBoard(mouseX, mouseY);
    setDragging({ itemId, offsetX: boardCoords.x - item.x, offsetY: boardCoords.y - item.y });
  }, [isConnecting, items, screenToBoard]);

  const onCanvasMouseMove = useCallback((e: React.MouseEvent) => {
    if (isPanning && !dragging.itemId) {
      // Handle panning
      const deltaX = e.clientX - panStart.x;
      const deltaY = e.clientY - panStart.y;
      setPanX(panStart.panX + deltaX);
      setPanY(panStart.panY + deltaY);
      return;
    }
    
    if (!dragging.itemId || !canvasRef.current) return;
    const rect = canvasRef.current.getBoundingClientRect();
    const mouseX = e.clientX - rect.left;
    const mouseY = e.clientY - rect.top;
    const boardCoords = screenToBoard(mouseX, mouseY);
    const x = boardCoords.x - dragging.offsetX;
    const y = boardCoords.y - dragging.offsetY;
    setItems((prev) => prev.map((it) => (it.id === dragging.itemId ? { ...it, x, y } : it)));
  }, [dragging, isPanning, panStart, screenToBoard]);

  const onCanvasMouseUp = useCallback(() => {
    if (isPanning) {
      setIsPanning(false);
    }
    if (!dragging.itemId) return;
    const moved = items.find((it) => it.id === dragging.itemId);
    setDragging({ itemId: null, offsetX: 0, offsetY: 0 });
    if (!moved || !moved.serverId || !id) return;
    // Persist position best-effort
    boardsApi.updateBoardItem(id, moved.serverId, { x: moved.x, y: moved.y }).catch(() => {});
  }, [dragging.itemId, id, items, isPanning]);

  const onItemClick = useCallback((itemId: string, isEditableClick = false) => {
    if (isConnecting) {
      // Handle connection mode
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
            setConnections((prev) => {
              const updated = prev.map((c) => (c.id === localId ? { 
                id: created.id, 
                from_item_id: firstConnectId, // Keep frontend IDs for rendering
                to_item_id: itemId 
              } : c));
              return updated;
            });
          })
          .catch((error) => {
            // Revert local if failed
            setConnections((prev) => prev.filter((c) => c.id !== localId));
          });
      }
    } else if (!isEditableClick) {
      // Handle selection mode (only when not clicking on editable fields)
      setSelectedItems(prev => {
        const newSelection = new Set(prev);
        if (newSelection.has(itemId)) {
          newSelection.delete(itemId);
        } else {
          newSelection.add(itemId);
        }
        return newSelection;
      });
    }
  }, [firstConnectId, isConnecting, items, id]);

  const handleUpdateContent = useCallback((itemId: string, newContent: string) => {
    setItems((prev) => prev.map((it) => it.id === itemId ? { ...it, content: newContent } : it));
    const item = items.find((it) => it.id === itemId);
    if (!item || !item.serverId || !id) return;
    boardsApi.updateBoardItem(id, item.serverId, { content: newContent }).catch(() => {});
  }, [id, items]);

  // Pan and zoom event handlers
  const onCanvasMouseDown = useCallback((e: React.MouseEvent) => {
    if (e.button === 1 || (e.button === 0 && e.ctrlKey)) { // Middle mouse or Ctrl+Left click for panning
      e.preventDefault();
      setIsPanning(true);
      setPanStart({
        x: e.clientX,
        y: e.clientY,
        panX,
        panY
      });
    }
  }, [panX, panY]);

  const onCanvasWheel = useCallback((e: React.WheelEvent) => {
    // Zoom is now handled by the global wheel handler
    // This handler can be used for other canvas-specific wheel behaviors if needed
    if (!e.ctrlKey && !e.metaKey) {
      // Allow normal scrolling when Ctrl/Cmd is not pressed
      // (though the canvas typically doesn't scroll)
    }
  }, []);

  // Keyboard shortcuts and global wheel handler
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Don't handle keys if user is typing in an input
      if ((e.target as HTMLElement)?.tagName === 'INPUT' || 
          (e.target as HTMLElement)?.tagName === 'TEXTAREA') {
        return;
      }

      if (e.ctrlKey || e.metaKey) {
        switch (e.key) {
          case '=':
          case '+':
            e.preventDefault();
            zoomIn();
            break;
          case '-':
            e.preventDefault();
            zoomOut();
            break;
          case '0':
            e.preventDefault();
            resetZoom();
            break;
        }
      } else if (e.key === 'Delete' || e.key === 'Backspace') {
        e.preventDefault();
        handleDeleteSelected();
      }
    };

    const handleGlobalWheel = (e: WheelEvent) => {
      if (e.ctrlKey || e.metaKey) {
        // Always prevent the default browser zoom behavior
        e.preventDefault();
        e.stopPropagation();
        
        // Apply custom board zoom if mouse is over the board canvas
        const target = e.target as HTMLElement;
        if (canvasRef.current && canvasRef.current.contains(target)) {
          const delta = e.deltaY > 0 ? -0.01 : 0.01; // 1% zoom increments
          setZoom(prev => Math.max(0.1, Math.min(3.0, prev + delta)));
        }
        // If not over the board, we've prevented browser zoom but don't apply any custom zoom
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    window.addEventListener('wheel', handleGlobalWheel, { passive: false });
    
    return () => {
      window.removeEventListener('keydown', handleKeyDown);
      window.removeEventListener('wheel', handleGlobalWheel);
    };
  }, [zoomIn, zoomOut, resetZoom, handleDeleteSelected, setZoom]);

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
          backgroundColor: '#c4a484', // Fallback solid color
          cursor: isConnecting ? 'crosshair' : (isPanning ? 'grabbing' : 'grab'),
          overflow: 'hidden',
        }}
        data-testid="board-canvas"
        onMouseDown={onCanvasMouseDown}
        onMouseMove={onCanvasMouseMove}
        onMouseUp={onCanvasMouseUp}
        onWheel={onCanvasWheel}
      >
        {/* Zoomable board content container */}
        <Box
          sx={{
            position: 'absolute',
            top: 0,
            left: 0,
            width: '400vw', // 4x larger board
            height: '400vh', // 4x larger board
            transform: `translate(${panX}px, ${panY}px) scale(${zoom})`,
            transformOrigin: '0 0',
            background: `
              radial-gradient(circle at 20% 30%, rgba(139, 119, 101, 0.8), transparent 2px),
              radial-gradient(circle at 70% 20%, rgba(160, 142, 124, 0.6), transparent 2px),
              radial-gradient(circle at 40% 70%, rgba(120, 100, 85, 0.7), transparent 2px),
              radial-gradient(circle at 90% 80%, rgba(145, 125, 108, 0.5), transparent 2px),
              linear-gradient(45deg, #c4a484 0%, #d4b896 25%, #c8a688 50%, #ddc2a4 75%, #c4a484 100%)
            `,
            backgroundSize: '50px 50px, 75px 75px, 60px 60px, 40px 40px, 100% 100%',
          }}
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
              onClick={(isEditableClick) => onItemClick(item.id, isEditableClick)}
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
      </Box>

      {/* Zoom Controls - Fixed position in top-left */}
      <Paper
        elevation={2}
        sx={{
          position: 'absolute',
          top: 90,
          right: 16,
          zIndex: 1100,
          p: 1,
          display: 'flex',
          flexDirection: 'column',
          gap: 1,
          bgcolor: 'background.paper',
        }}
      >
        <IconButton
          size="small"
          onClick={zoomIn}
          title="Zoom In (Ctrl/Cmd + + or Ctrl/Cmd + Mouse Wheel)"
          sx={{ fontSize: '0.8rem' }}
        >
          <ZoomInIcon fontSize="small" />
        </IconButton>
        <Typography variant="caption" align="center" sx={{ px: 0.5, minWidth: '40px' }}>
          {Math.round(zoom * 100)}%
        </Typography>
        <IconButton
          size="small"
          onClick={zoomOut}
          title="Zoom Out (Ctrl/Cmd + - or Ctrl/Cmd + Mouse Wheel)"
          sx={{ fontSize: '0.8rem' }}
        >
          <ZoomOutIcon fontSize="small" />
        </IconButton>
        <IconButton
          size="small"
          onClick={resetZoom}
          title="Reset Zoom (Ctrl/Cmd + 0)"
          sx={{ fontSize: '0.8rem' }}
        >
          <ResetZoomIcon fontSize="small" />
        </IconButton>
      </Paper>

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
          <SpeedDialAction
            icon={<DeleteIcon />}
            tooltipTitle={`Delete Selected (${selectedItems.size} items)`}
            onClick={selectedItems.size > 0 ? handleDeleteSelected : undefined}
            sx={{ 
              bgcolor: selectedItems.size > 0 ? 'error.main' : 'action.disabled',
              '&:hover': {
                bgcolor: selectedItems.size > 0 ? 'error.dark' : 'action.disabled'
              },
              opacity: selectedItems.size > 0 ? 1 : 0.5,
              pointerEvents: selectedItems.size > 0 ? 'auto' : 'none'
            }}
            data-testid="delete-selected-button"
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
  onClick: (isEditableClick?: boolean) => void;
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

  const handleClick = React.useCallback((e: React.MouseEvent, isEditableClick = false) => {
    e.stopPropagation();
    onClick(isEditableClick);
  }, [onClick]);

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
        boxShadow: isSelected
          ? (isPostIt 
              ? '0 4px 8px rgba(0,0,0,0.2), inset 0 0 0 1px rgba(0,0,0,0.1), 0 0 0 3px rgba(25,118,210,0.3)'
              : '0 6px 12px rgba(0,0,0,0.3), inset 0 0 0 1px rgba(0,0,0,0.1), 0 0 0 3px rgba(25,118,210,0.3)')
          : (isPostIt 
              ? '0 4px 8px rgba(0,0,0,0.2), inset 0 0 0 1px rgba(0,0,0,0.1)'
              : '0 6px 12px rgba(0,0,0,0.3), inset 0 0 0 1px rgba(0,0,0,0.1)'),
        p: isPostIt ? '20px 15px 15px 15px' : 2,
        '&:hover': canEdit ? {
          transform: `rotate(${item.rotation}deg) scale(1.02)`,
          zIndex: 100,
        } : {},
      }}
      onMouseDown={canEdit ? onMouseDown : undefined}
      onClick={(e) => handleClick(e, false)}
      data-testid="board-item"
      onDoubleClick={() => {
        if (!canEdit) return;
        // If content is just the placeholder, clear it when entering edit mode
        const trimmed = (item.content || '').trim();
        if (isPostIt && (!trimmed || trimmed === 'Evidence notes...')) {
          setDraft('');
        }
        setIsEditing(true);
      }}
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
      {isPostIt ? (
        isEditing ? (
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
                if (!e.shiftKey) {
                  e.preventDefault();
                  commit();
                }
              } else if (e.key === 'Enter' && !isPostIt) {
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
            placeholder={isPostIt ? '' : undefined}
            sx={{
              '& .MuiInputBase-input': {
                fontFamily: '"Courier New", monospace',
                fontSize: 14,
                lineHeight: 1.4,
                color: '#333',
                whiteSpace: 'pre-wrap',
                padding: 0,
              },
              backgroundColor: 'rgba(255,255,255,0.9)'
            }}
          />
        ) : (
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
        )
      ) : (
          <Box>
            {/* Suspect card photo area */}
            <Box
              sx={{
                width: '100%',
                height: 100,
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

            <SuspectCardContent
              content={item.content}
              canEdit={canEdit}
              isParentEditing={isEditing}
              onRequestSetParentEditing={setIsEditing}
              onUpdateContent={onUpdateContent}
              onEditableClick={(e) => handleClick(e, true)}
            />
          </Box>
      )}
    </Box>
  );
};

export default BoardPage;

// Child component to render and edit suspect card content without violating hook rules
interface SuspectCardContentProps {
  content: string;
  canEdit: boolean;
  isParentEditing: boolean;
  onRequestSetParentEditing: (editing: boolean) => void;
  onUpdateContent: (content: string) => void;
  onEditableClick: (e: React.MouseEvent) => void;
}

const SuspectCardContent: React.FC<SuspectCardContentProps> = ({
  content,
  canEdit,
  isParentEditing,
  onRequestSetParentEditing,
  onUpdateContent,
  onEditableClick,
}) => {
  const parse = (raw: string) => {
    const lines = (raw || '').split('\n');
    const name = (lines[0] || '').trim();
    const ageLine = lines.find((l) => l.trim().toLowerCase().startsWith('age:')) || 'Age:';
    const lastSeenLine = lines.find((l) => l.trim().toLowerCase().startsWith('last seen:')) || 'Last seen:';
    const notesIndex = lines.findIndex((l) => l.trim().toLowerCase().startsWith('notes:'));
    const age = ageLine.split(':').slice(1).join(':').trim();
    const lastSeen = lastSeenLine.split(':').slice(1).join(':').trim();
    const notes = notesIndex >= 0 ? lines.slice(notesIndex + 1).join('\n') : '';
    return { name, age, lastSeen, notes };
  };
  const format = (s: { name: string; age: string; lastSeen: string; notes: string }) => {
    const header = `${s.name || ''}`;
    const ageL = `Age: ${s.age || ''}`;
    const lastL = `Last seen: ${s.lastSeen || ''}`;
    const notesHeader = 'Notes:';
    const notesBody = (s.notes || '').length ? `\n${s.notes}` : '';
    return `${header}\n${ageL}\n${lastL}\n${notesHeader}${notesBody}`;
  };

  const stopDrag = (e: React.MouseEvent) => e.stopPropagation();
  const [suspect, setSuspect] = React.useState(parse(content || ''));
  const [suspectIsEditing, setSuspectIsEditing] = React.useState(false);
  const cardRef = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    setSuspect(parse(content || ''));
  }, [content]);

  React.useEffect(() => {
    setSuspectIsEditing(isParentEditing);
  }, [isParentEditing]);

  const commitSuspect = React.useCallback(() => {
    onUpdateContent(format(suspect));
    setSuspectIsEditing(false);
    onRequestSetParentEditing(false);
  }, [suspect, onUpdateContent, onRequestSetParentEditing, format]);

  const handleFieldBlur = React.useCallback((e: React.FocusEvent) => {
    // Check if the new focus target is still within the card
    const relatedTarget = e.relatedTarget as HTMLElement;
    if (cardRef.current && relatedTarget && cardRef.current.contains(relatedTarget)) {
      // Focus is moving to another field within the same card, don't commit
      return;
    }
    // Focus is leaving the card entirely, commit the changes
    commitSuspect();
  }, [commitSuspect]);

  const handleFieldClick = (e: React.MouseEvent) => {
    if (suspectIsEditing) {
      onEditableClick(e);
    }
  };

  // Set up click outside listener when editing
  React.useEffect(() => {
    if (!suspectIsEditing) return;

    const handleClickOutside = (event: MouseEvent) => {
      if (cardRef.current && !cardRef.current.contains(event.target as Node)) {
        commitSuspect();
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [suspectIsEditing, commitSuspect]);

  if (!canEdit || !suspectIsEditing) {
    return (
      <>
        <Typography variant="h6" sx={{ fontWeight: 'bold', mb: 1 }}>
          {suspect.name || 'Suspect Name'}
        </Typography>
        <Typography variant="body2" color="textSecondary" sx={{ whiteSpace: 'pre-wrap' }}>
          {`Age: ${suspect.age || 'Unknown'}\nLast seen: ${suspect.lastSeen || ''}\nNotes:`}
        </Typography>
        {suspect.notes && (
          <Typography 
            variant="body2" 
            sx={{ 
              whiteSpace: 'pre-wrap', 
              mt: 1,
              maxHeight: '120px',
              overflow: 'auto',
              padding: '4px',
              border: '1px solid #e0e0e0',
              borderRadius: '4px',
              backgroundColor: '#fafafa'
            }}
          >
            {suspect.notes}
          </Typography>
        )}
      </>
    );
  }

  return (
    <Box ref={cardRef}>
      <TextField
        label="Name"
        value={suspect.name}
        onChange={(e) => setSuspect((s) => ({ ...s, name: e.target.value }))}
        onBlur={handleFieldBlur}
        variant="outlined"
        size="small"
        fullWidth
        sx={{
          mb: 1,
          '& .MuiInputBase-input::placeholder': { opacity: 1 },
          '&.Mui-focused .MuiInputBase-input::placeholder': { opacity: 0 }
        }}
        onMouseDown={stopDrag}
        onClick={handleFieldClick}
        placeholder="Suspect Name"
        inputProps={{ 'data-testid': 'suspect-name-input' }}
      />
      <Box sx={{ display: 'flex', gap: 1, mb: 1 }}>
        <TextField
          label="Age"
          value={suspect.age}
          onChange={(e) => setSuspect((s) => ({ ...s, age: e.target.value }))}
          onBlur={handleFieldBlur}
          variant="outlined"
          size="small"
          sx={{
            width: 140,
            '& .MuiInputBase-input::placeholder': { opacity: 1 },
            '&.Mui-focused .MuiInputBase-input::placeholder': { opacity: 0 }
          }}
          onMouseDown={stopDrag}
          onClick={handleFieldClick}
          placeholder="Unknown"
          inputProps={{ 'data-testid': 'suspect-age-input' }}
        />
        <TextField
          label="Last seen"
          value={suspect.lastSeen}
          onChange={(e) => setSuspect((s) => ({ ...s, lastSeen: e.target.value }))}
          onBlur={handleFieldBlur}
          variant="outlined"
          size="small"
          fullWidth
          onMouseDown={stopDrag}
          onClick={handleFieldClick}
          sx={{
            '& .MuiInputBase-input::placeholder': { opacity: 1 },
            '&.Mui-focused .MuiInputBase-input::placeholder': { opacity: 0 }
          }}
          placeholder="Location / time"
          inputProps={{ 'data-testid': 'suspect-lastseen-input' }}
        />
      </Box>
      <TextField
        label="Notes"
        value={suspect.notes}
        onChange={(e) => setSuspect((s) => ({ ...s, notes: e.target.value }))}
        onBlur={handleFieldBlur}
        variant="outlined"
        size="small"
        fullWidth
        multiline
        minRows={3}
        maxRows={6}
        onMouseDown={stopDrag}
        onClick={handleFieldClick}
        sx={{
          '& .MuiInputBase-input::placeholder': { opacity: 1 },
          '&.Mui-focused .MuiInputBase-input::placeholder': { opacity: 0 }
        }}
        placeholder="Add notes..."
        inputProps={{ 'data-testid': 'suspect-notes-input' }}
      />
    </Box>
  );
};


