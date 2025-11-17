import React, { useState } from 'react';
import {
  Box,
  Paper,
  Typography,
  Button,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  IconButton,
  Chip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Grid
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  Upload as UploadIcon,
  Download as DownloadIcon,
  Visibility as ViewIcon
} from '@mui/icons-material';

/**
 * Contact Lists Page
 * Create, Import, Export contact groups
 */
function ContactLists() {
  const [lists, setLists] = useState([
    {
      id: 1,
      name: 'Marketing List',
      description: 'Active marketing subscribers',
      contacts_count: 5234,
      tags: ['marketing', 'active'],
      created_at: '2025-01-10',
      last_updated: '2025-01-15'
    },
    {
      id: 2,
      name: 'Premium Customers',
      description: 'High-value customer segment',
      contacts_count: 1567,
      tags: ['premium', 'vip'],
      created_at: '2025-01-05',
      last_updated: '2025-01-14'
    },
    {
      id: 3,
      name: 'Inactive Users',
      description: 'Users inactive for 30+ days',
      contacts_count: 8901,
      tags: ['inactive', 'churn-risk'],
      created_at: '2025-01-12',
      last_updated: '2025-01-15'
    },
  ]);

  const [openDialog, setOpenDialog] = useState(false);
  const [openImportDialog, setOpenImportDialog] = useState(false);
  const [selectedList, setSelectedList] = useState(null);
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    tags: ''
  });

  const handleOpenDialog = (list = null) => {
    if (list) {
      setSelectedList(list);
      setFormData({
        name: list.name,
        description: list.description,
        tags: list.tags.join(', ')
      });
    } else {
      setSelectedList(null);
      setFormData({
        name: '',
        description: '',
        tags: ''
      });
    }
    setOpenDialog(true);
  };

  const handleCloseDialog = () => {
    setOpenDialog(false);
    setSelectedList(null);
  };

  const handleSave = () => {
    console.log('Saving contact list:', formData);
    handleCloseDialog();
  };

  const handleDelete = (id) => {
    if (window.confirm('Are you sure you want to delete this contact list?')) {
      console.log('Deleting contact list:', id);
    }
  };

  const handleImport = () => {
    setOpenImportDialog(true);
  };

  const handleExport = (id) => {
    console.log('Exporting contact list:', id);
    // Implement export logic
  };

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h5">Contact Lists</Typography>
        <Box>
          <Button
            variant="outlined"
            startIcon={<UploadIcon />}
            onClick={handleImport}
            sx={{ mr: 1 }}
          >
            Import Contacts
          </Button>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => handleOpenDialog()}
          >
            Create List
          </Button>
        </Box>
      </Box>

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>List Name</TableCell>
              <TableCell>Description</TableCell>
              <TableCell>Contacts</TableCell>
              <TableCell>Tags</TableCell>
              <TableCell>Last Updated</TableCell>
              <TableCell align="right">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {lists.map((list) => (
              <TableRow key={list.id}>
                <TableCell>
                  <Typography variant="body2" fontWeight="bold">
                    {list.name}
                  </Typography>
                </TableCell>
                <TableCell>{list.description}</TableCell>
                <TableCell>
                  <Chip
                    label={list.contacts_count.toLocaleString()}
                    color="primary"
                    size="small"
                  />
                </TableCell>
                <TableCell>
                  {list.tags.map(tag => (
                    <Chip key={tag} label={tag} size="small" sx={{ mr: 0.5 }} />
                  ))}
                </TableCell>
                <TableCell>{list.last_updated}</TableCell>
                <TableCell align="right">
                  <IconButton size="small" onClick={() => console.log('View', list.id)}>
                    <ViewIcon />
                  </IconButton>
                  <IconButton size="small" onClick={() => handleOpenDialog(list)}>
                    <EditIcon />
                  </IconButton>
                  <IconButton size="small" onClick={() => handleExport(list.id)}>
                    <DownloadIcon />
                  </IconButton>
                  <IconButton size="small" color="error" onClick={() => handleDelete(list.id)}>
                    <DeleteIcon />
                  </IconButton>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>

      {/* Create/Edit Dialog */}
      <Dialog open={openDialog} onClose={handleCloseDialog} maxWidth="sm" fullWidth>
        <DialogTitle>
          {selectedList ? 'Edit Contact List' : 'Create New Contact List'}
        </DialogTitle>
        <DialogContent>
          <Grid container spacing={2} sx={{ mt: 1 }}>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="List Name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              />
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="Description"
                multiline
                rows={3}
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              />
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="Tags"
                value={formData.tags}
                onChange={(e) => setFormData({ ...formData, tags: e.target.value })}
                helperText="Comma-separated tags"
              />
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDialog}>Cancel</Button>
          <Button variant="contained" onClick={handleSave}>
            {selectedList ? 'Update' : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>

      {/* Import Dialog */}
      <Dialog open={openImportDialog} onClose={() => setOpenImportDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>Import Contacts</DialogTitle>
        <DialogContent>
          <Box sx={{ mt: 2 }}>
            <Typography variant="body2" gutterBottom>
              Upload a CSV, TXT, or Excel file with phone numbers
            </Typography>
            <Button
              variant="outlined"
              component="label"
              startIcon={<UploadIcon />}
              fullWidth
              sx={{ mt: 2 }}
            >
              Select File
              <input
                type="file"
                hidden
                accept=".csv,.txt,.xlsx"
              />
            </Button>
            <Typography variant="caption" color="textSecondary" sx={{ mt: 2, display: 'block' }}>
              Supports up to 1M records per file
            </Typography>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenImportDialog(false)}>Cancel</Button>
          <Button variant="contained">Import</Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}

export default ContactLists;
