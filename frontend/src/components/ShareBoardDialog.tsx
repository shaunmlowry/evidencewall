import { Share as ShareIcon } from '@mui/icons-material';
import {
    Alert,
    Box,
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogTitle,
    FormControl,
    InputLabel,
    MenuItem,
    Select,
    TextField,
    Typography,
} from '@mui/material';
import { useMutation } from '@tanstack/react-query';
import React, { useState } from 'react';
import { boardsApi } from '../services/api';
import { Board, PermissionLevel } from '../types';

interface ShareBoardDialogProps {
    open: boolean;
    board: Board | null;
    onClose: () => void;
    onSuccess?: () => void;
}

const ShareBoardDialog: React.FC<ShareBoardDialogProps> = ({ open, board, onClose, onSuccess }) => {
    const [email, setEmail] = useState('');
    const [permission, setPermission] = useState<PermissionLevel>('read');
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');

    const shareMutation = useMutation({
        mutationFn: (data: { user_email: string; permission: PermissionLevel }) => {
            if (!board) throw new Error('No board selected');
            return boardsApi.shareBoard(board.id, data);
        },
        onSuccess: () => {
            setSuccess(`Board shared successfully with ${email}`);
            setEmail('');
            setPermission('read');
            setError('');
            setTimeout(() => {
                setSuccess('');
                if (onSuccess) onSuccess();
                onClose();
            }, 2000);
        },
        onError: (err: any) => {
            const errorMessage = err.response?.data?.error || 'Failed to share board';
            setError(errorMessage);
            setSuccess('');
        },
    });

    const handleShare = () => {
        setError('');
        setSuccess('');

        // Validate email
        const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        if (!email.trim()) {
            setError('Please enter an email address');
            return;
        }
        if (!emailRegex.test(email)) {
            setError('Please enter a valid email address');
            return;
        }

        shareMutation.mutate({
            user_email: email,
            permission,
        });
    };

    const handleClose = () => {
        if (!shareMutation.isPending) {
            setEmail('');
            setPermission('read');
            setError('');
            setSuccess('');
            onClose();
        }
    };

    return (
        <Dialog open={open} onClose={handleClose} maxWidth="sm" fullWidth>
            <DialogTitle>
                <Box display="flex" alignItems="center" gap={1}>
                    <ShareIcon />
                    Share Board: {board?.title}
                </Box>
            </DialogTitle>
            <DialogContent>
                {error && (
                    <Alert severity="error" sx={{ mb: 2 }}>
                        {error}
                    </Alert>
                )}
                {success && (
                    <Alert severity="success" sx={{ mb: 2 }}>
                        {success}
                    </Alert>
                )}

                <Typography variant="body2" color="textSecondary" mb={2}>
                    Enter the email address of a registered user to share this board with them.
                </Typography>

                <TextField
                    autoFocus
                    margin="dense"
                    label="User Email Address"
                    type="email"
                    fullWidth
                    variant="outlined"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    onKeyPress={(e) => {
                        if (e.key === 'Enter') {
                            e.preventDefault();
                            handleShare();
                        }
                    }}
                    placeholder="user@example.com"
                    sx={{ mb: 2 }}
                    disabled={shareMutation.isPending}
                />

                <FormControl fullWidth>
                    <InputLabel>Permission Level</InputLabel>
                    <Select
                        value={permission}
                        label="Permission Level"
                        onChange={(e) => setPermission(e.target.value as PermissionLevel)}
                        disabled={shareMutation.isPending}
                    >
                        <MenuItem value="read">
                            <Box>
                                <Typography variant="body1">Read Only</Typography>
                                <Typography variant="caption" color="textSecondary">
                                    Can view the board
                                </Typography>
                            </Box>
                        </MenuItem>
                        <MenuItem value="write">
                            <Box>
                                <Typography variant="body1">Write</Typography>
                                <Typography variant="caption" color="textSecondary">
                                    Can view and edit the board
                                </Typography>
                            </Box>
                        </MenuItem>
                        <MenuItem value="admin">
                            <Box>
                                <Typography variant="body1">Admin</Typography>
                                <Typography variant="caption" color="textSecondary">
                                    Full control including sharing and deletion
                                </Typography>
                            </Box>
                        </MenuItem>
                    </Select>
                </FormControl>
            </DialogContent>
            <DialogActions>
                <Button onClick={handleClose} disabled={shareMutation.isPending}>
                    Cancel
                </Button>
                <Button
                    onClick={handleShare}
                    variant="contained"
                    disabled={shareMutation.isPending || !email.trim()}
                    startIcon={<ShareIcon />}
                >
                    {shareMutation.isPending ? 'Sharing...' : 'Share'}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default ShareBoardDialog;

