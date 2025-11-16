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
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Grid
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  FileCopy as CopyIcon
} from '@mui/icons-material';

/**
 * Message Templates Page
 * Create/Edit/Delete message templates with multi-language support
 */
function MessageTemplates() {
  const [templates, setTemplates] = useState([
    {
      id: 1,
      name: 'OTP Verification',
      content: 'Your verification code is %CODE%. Valid for 5 minutes.',
      category: 'OTP',
      language: 'EN',
      variables: ['CODE'],
      visibility: 'PUBLIC',
      created_by: 'admin'
    },
    {
      id: 2,
      name: 'Welcome Message',
      content: 'Welcome to %COMPANY%! We are glad to have you.',
      category: 'MARKETING',
      language: 'EN',
      variables: ['COMPANY'],
      visibility: 'PUBLIC',
      created_by: 'admin'
    },
    {
      id: 3,
      name: 'Balance Alert',
      content: 'Your account balance is %BALANCE%. Please recharge.',
      category: 'ALERT',
      language: 'EN',
      variables: ['BALANCE'],
      visibility: 'RESELLER',
      created_by: 'reseller1'
    },
  ]);

  const [openDialog, setOpenDialog] = useState(false);
  const [selectedTemplate, setSelectedTemplate] = useState(null);
  const [formData, setFormData] = useState({
    name: '',
    content: '',
    category: 'MARKETING',
    language: 'EN',
    visibility: 'PUBLIC'
  });

  const handleOpenDialog = (template = null) => {
    if (template) {
      setSelectedTemplate(template);
      setFormData(template);
    } else {
      setSelectedTemplate(null);
      setFormData({
        name: '',
        content: '',
        category: 'MARKETING',
        language: 'EN',
        visibility: 'PUBLIC'
      });
    }
    setOpenDialog(true);
  };

  const handleCloseDialog = () => {
    setOpenDialog(false);
    setSelectedTemplate(null);
  };

  const handleSave = () => {
    console.log('Saving template:', formData);
    handleCloseDialog();
  };

  const handleDelete = (id) => {
    if (window.confirm('Are you sure you want to delete this template?')) {
      console.log('Deleting template:', id);
    }
  };

  const getCategoryColor = (category) => {
    const colors = {
      OTP: 'error',
      MARKETING: 'primary',
      ALERT: 'warning',
      INFO: 'info'
    };
    return colors[category] || 'default';
  };

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h5">Message Templates</Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={() => handleOpenDialog()}
        >
          Create Template
        </Button>
      </Box>

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Template Name</TableCell>
              <TableCell>Content</TableCell>
              <TableCell>Category</TableCell>
              <TableCell>Language</TableCell>
              <TableCell>Variables</TableCell>
              <TableCell>Visibility</TableCell>
              <TableCell align="right">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {templates.map((template) => (
              <TableRow key={template.id}>
                <TableCell>
                  <Typography variant="body2" fontWeight="bold">
                    {template.name}
                  </Typography>
                </TableCell>
                <TableCell sx={{ maxWidth: 300 }}>
                  <Typography variant="body2" noWrap>
                    {template.content}
                  </Typography>
                </TableCell>
                <TableCell>
                  <Chip
                    label={template.category}
                    color={getCategoryColor(template.category)}
                    size="small"
                  />
                </TableCell>
                <TableCell>{template.language}</TableCell>
                <TableCell>
                  {template.variables && template.variables.map(v => (
                    <Chip key={v} label={`%${v}%`} size="small" sx={{ mr: 0.5 }} />
                  ))}
                </TableCell>
                <TableCell>
                  <Chip label={template.visibility} size="small" variant="outlined" />
                </TableCell>
                <TableCell align="right">
                  <IconButton size="small" onClick={() => handleOpenDialog(template)}>
                    <EditIcon />
                  </IconButton>
                  <IconButton size="small" onClick={() => console.log('Duplicate', template.id)}>
                    <CopyIcon />
                  </IconButton>
                  <IconButton size="small" color="error" onClick={() => handleDelete(template.id)}>
                    <DeleteIcon />
                  </IconButton>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>

      {/* Create/Edit Dialog */}
      <Dialog open={openDialog} onClose={handleCloseDialog} maxWidth="md" fullWidth>
        <DialogTitle>
          {selectedTemplate ? 'Edit Template' : 'Create New Template'}
        </DialogTitle>
        <DialogContent>
          <Grid container spacing={2} sx={{ mt: 1 }}>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="Template Name"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              />
            </Grid>
            <Grid item xs={12}>
              <TextField
                fullWidth
                label="Message Content"
                multiline
                rows={4}
                value={formData.content}
                onChange={(e) => setFormData({ ...formData, content: e.target.value })}
                helperText="Use %VARIABLE% for dynamic content (e.g., %NAME%, %CODE%, %BALANCE%)"
              />
            </Grid>
            <Grid item xs={12} sm={4}>
              <FormControl fullWidth>
                <InputLabel>Category</InputLabel>
                <Select
                  value={formData.category}
                  label="Category"
                  onChange={(e) => setFormData({ ...formData, category: e.target.value })}
                >
                  <MenuItem value="OTP">OTP</MenuItem>
                  <MenuItem value="MARKETING">Marketing</MenuItem>
                  <MenuItem value="ALERT">Alert</MenuItem>
                  <MenuItem value="INFO">Info</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} sm={4}>
              <FormControl fullWidth>
                <InputLabel>Language</InputLabel>
                <Select
                  value={formData.language}
                  label="Language"
                  onChange={(e) => setFormData({ ...formData, language: e.target.value })}
                >
                  <MenuItem value="EN">English</MenuItem>
                  <MenuItem value="AR">Arabic</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} sm={4}>
              <FormControl fullWidth>
                <InputLabel>Visibility</InputLabel>
                <Select
                  value={formData.visibility}
                  label="Visibility"
                  onChange={(e) => setFormData({ ...formData, visibility: e.target.value })}
                >
                  <MenuItem value="PUBLIC">Public</MenuItem>
                  <MenuItem value="RESELLER">Reseller Only</MenuItem>
                  <MenuItem value="USER">User Only</MenuItem>
                </Select>
              </FormControl>
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDialog}>Cancel</Button>
          <Button variant="contained" onClick={handleSave}>
            {selectedTemplate ? 'Update' : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}

export default MessageTemplates;
