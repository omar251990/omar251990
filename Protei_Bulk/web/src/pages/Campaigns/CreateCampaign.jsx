import React, { useState } from 'react';
import {
  Box,
  Paper,
  Typography,
  Stepper,
  Step,
  StepLabel,
  Button,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Grid,
  RadioGroup,
  FormControlLabel,
  Radio,
  Chip,
  Alert,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  IconButton
} from '@mui/material';
import {
  Upload as UploadIcon,
  Delete as DeleteIcon,
  Add as AddIcon
} from '@mui/icons-material';

/**
 * Create Campaign Page
 * Multi-step wizard for campaign creation
 */
function CreateCampaign() {
  const [activeStep, setActiveStep] = useState(0);
  const [campaignData, setCampaignData] = useState({
    // Step 1: Basic Info
    name: '',
    channel: 'SMS',
    sender_id: '',
    priority: 'NORMAL',

    // Step 2: Recipients
    recipient_type: 'FILE',
    file: null,
    contact_list_id: '',
    profile_group_id: '',
    manual_msisdns: '',

    // Step 3: Message
    use_template: false,
    template_id: '',
    message_content: '',
    encoding: 'GSM7',
    variables: [],

    // Step 4: Schedule
    schedule_type: 'IMMEDIATE',
    schedule_date: '',
    schedule_time: '',
    max_messages_per_day: '',

    // Step 5: Review
    total_recipients: 0,
    estimated_cost: 0,
  });

  const steps = [
    'Channel & Sender',
    'Select Recipients',
    'Compose Message',
    'Schedule & Priority',
    'Review & Submit'
  ];

  const handleNext = () => {
    setActiveStep((prevActiveStep) => prevActiveStep + 1);
  };

  const handleBack = () => {
    setActiveStep((prevActiveStep) => prevActiveStep - 1);
  };

  const handleSubmit = async () => {
    console.log('Submitting campaign:', campaignData);
    // Submit campaign logic here
    alert('Campaign created successfully!');
  };

  const handleFileUpload = (event) => {
    const file = event.target.files[0];
    if (file) {
      setCampaignData({ ...campaignData, file });
      // Parse file to get total recipients
      setCampaignData({ ...campaignData, file, total_recipients: 1000 }); // Mock
    }
  };

  // Step 1: Channel & Sender
  const renderStep1 = () => (
    <Grid container spacing={3}>
      <Grid item xs={12}>
        <TextField
          fullWidth
          label="Campaign Name"
          value={campaignData.name}
          onChange={(e) => setCampaignData({ ...campaignData, name: e.target.value })}
          required
        />
      </Grid>
      <Grid item xs={12} sm={6}>
        <FormControl fullWidth>
          <InputLabel>Channel</InputLabel>
          <Select
            value={campaignData.channel}
            label="Channel"
            onChange={(e) => setCampaignData({ ...campaignData, channel: e.target.value })}
          >
            <MenuItem value="SMS">SMS</MenuItem>
            <MenuItem value="USSD">USSD</MenuItem>
            <MenuItem value="EMAIL">Email</MenuItem>
            <MenuItem value="PUSH">Push Notification</MenuItem>
            <MenuItem value="WHATSAPP">WhatsApp</MenuItem>
            <MenuItem value="TELEGRAM">Telegram</MenuItem>
          </Select>
        </FormControl>
      </Grid>
      <Grid item xs={12} sm={6}>
        <TextField
          fullWidth
          label="Sender ID"
          value={campaignData.sender_id}
          onChange={(e) => setCampaignData({ ...campaignData, sender_id: e.target.value })}
          helperText="Alphanumeric or numeric sender"
          required
        />
      </Grid>
      <Grid item xs={12} sm={6}>
        <FormControl fullWidth>
          <InputLabel>Priority</InputLabel>
          <Select
            value={campaignData.priority}
            label="Priority"
            onChange={(e) => setCampaignData({ ...campaignData, priority: e.target.value })}
          >
            <MenuItem value="CRITICAL">Critical</MenuItem>
            <MenuItem value="HIGH">High</MenuItem>
            <MenuItem value="NORMAL">Normal</MenuItem>
            <MenuItem value="LOW">Low</MenuItem>
          </Select>
        </FormControl>
      </Grid>
    </Grid>
  );

  // Step 2: Recipients
  const renderStep2 = () => (
    <Grid container spacing={3}>
      <Grid item xs={12}>
        <Typography variant="subtitle1" gutterBottom>
          Select Recipient Source
        </Typography>
        <RadioGroup
          value={campaignData.recipient_type}
          onChange={(e) => setCampaignData({ ...campaignData, recipient_type: e.target.value })}
        >
          <FormControlLabel value="FILE" control={<Radio />} label="Upload File (CSV/TXT/Excel)" />
          <FormControlLabel value="CONTACT_LIST" control={<Radio />} label="Select from Contact List" />
          <FormControlLabel value="PROFILE" control={<Radio />} label="Select Profile Group" />
          <FormControlLabel value="MANUAL" control={<Radio />} label="Manual MSISDN Entry" />
        </RadioGroup>
      </Grid>

      {campaignData.recipient_type === 'FILE' && (
        <Grid item xs={12}>
          <Button
            variant="outlined"
            component="label"
            startIcon={<UploadIcon />}
            fullWidth
          >
            Upload Recipients File
            <input
              type="file"
              hidden
              accept=".csv,.txt,.xlsx"
              onChange={handleFileUpload}
            />
          </Button>
          {campaignData.file && (
            <Alert severity="success" sx={{ mt: 2 }}>
              File uploaded: {campaignData.file.name} ({campaignData.total_recipients} recipients)
            </Alert>
          )}
        </Grid>
      )}

      {campaignData.recipient_type === 'CONTACT_LIST' && (
        <Grid item xs={12}>
          <FormControl fullWidth>
            <InputLabel>Contact List</InputLabel>
            <Select
              value={campaignData.contact_list_id}
              label="Contact List"
              onChange={(e) => setCampaignData({ ...campaignData, contact_list_id: e.target.value })}
            >
              <MenuItem value="1">Marketing List (5,234 contacts)</MenuItem>
              <MenuItem value="2">Premium Customers (1,567 contacts)</MenuItem>
              <MenuItem value="3">Inactive Users (8,901 contacts)</MenuItem>
            </Select>
          </FormControl>
        </Grid>
      )}

      {campaignData.recipient_type === 'PROFILE' && (
        <Grid item xs={12}>
          <FormControl fullWidth>
            <InputLabel>Profile Group</InputLabel>
            <Select
              value={campaignData.profile_group_id}
              label="Profile Group"
              onChange={(e) => setCampaignData({ ...campaignData, profile_group_id: e.target.value })}
            >
              <MenuItem value="1">Region=Amman AND Plan=Prepaid (12,345 users)</MenuItem>
              <MenuItem value="2">Age 18-25 AND Gender=Male (7,890 users)</MenuItem>
              <MenuItem value="3">High Value Customers (2,456 users)</MenuItem>
            </Select>
          </FormControl>
        </Grid>
      )}

      {campaignData.recipient_type === 'MANUAL' && (
        <Grid item xs={12}>
          <TextField
            fullWidth
            label="MSISDNs"
            multiline
            rows={6}
            value={campaignData.manual_msisdns}
            onChange={(e) => setCampaignData({ ...campaignData, manual_msisdns: e.target.value })}
            helperText="Enter phone numbers separated by comma or newline"
          />
        </Grid>
      )}
    </Grid>
  );

  // Step 3: Message Content
  const renderStep3 = () => (
    <Grid container spacing={3}>
      <Grid item xs={12}>
        <FormControlLabel
          control={
            <Radio
              checked={campaignData.use_template}
              onChange={(e) => setCampaignData({ ...campaignData, use_template: e.target.checked })}
            />
          }
          label="Use Message Template"
        />
        <FormControlLabel
          control={
            <Radio
              checked={!campaignData.use_template}
              onChange={(e) => setCampaignData({ ...campaignData, use_template: !e.target.checked })}
            />
          }
          label="Write Custom Message"
        />
      </Grid>

      {campaignData.use_template ? (
        <Grid item xs={12}>
          <FormControl fullWidth>
            <InputLabel>Select Template</InputLabel>
            <Select
              value={campaignData.template_id}
              label="Select Template"
              onChange={(e) => setCampaignData({ ...campaignData, template_id: e.target.value })}
            >
              <MenuItem value="1">OTP Template - Your code is %CODE%</MenuItem>
              <MenuItem value="2">Welcome Message - Welcome to %COMPANY%</MenuItem>
              <MenuItem value="3">Promo Alert - Special offer %DISCOUNT%</MenuItem>
            </Select>
          </FormControl>
        </Grid>
      ) : (
        <Grid item xs={12}>
          <TextField
            fullWidth
            label="Message Content"
            multiline
            rows={6}
            value={campaignData.message_content}
            onChange={(e) => setCampaignData({ ...campaignData, message_content: e.target.value })}
            helperText={`Characters: ${campaignData.message_content.length} | Parts: ${Math.ceil(campaignData.message_content.length / 160)}`}
          />
        </Grid>
      )}

      <Grid item xs={12} sm={6}>
        <FormControl fullWidth>
          <InputLabel>Encoding</InputLabel>
          <Select
            value={campaignData.encoding}
            label="Encoding"
            onChange={(e) => setCampaignData({ ...campaignData, encoding: e.target.value })}
          >
            <MenuItem value="GSM7">GSM7 (160 chars/part)</MenuItem>
            <MenuItem value="UCS2">UCS2 / Unicode (70 chars/part)</MenuItem>
            <MenuItem value="ASCII">ASCII</MenuItem>
          </Select>
        </FormControl>
      </Grid>
    </Grid>
  );

  // Step 4: Schedule
  const renderStep4 = () => (
    <Grid container spacing={3}>
      <Grid item xs={12}>
        <Typography variant="subtitle1" gutterBottom>
          Schedule Type
        </Typography>
        <RadioGroup
          value={campaignData.schedule_type}
          onChange={(e) => setCampaignData({ ...campaignData, schedule_type: e.target.value })}
        >
          <FormControlLabel value="IMMEDIATE" control={<Radio />} label="Send Immediately" />
          <FormControlLabel value="SCHEDULED" control={<Radio />} label="Schedule for Later" />
          <FormControlLabel value="RECURRING" control={<Radio />} label="Recurring Campaign" />
        </RadioGroup>
      </Grid>

      {campaignData.schedule_type === 'SCHEDULED' && (
        <>
          <Grid item xs={12} sm={6}>
            <TextField
              fullWidth
              label="Schedule Date"
              type="date"
              value={campaignData.schedule_date}
              onChange={(e) => setCampaignData({ ...campaignData, schedule_date: e.target.value })}
              InputLabelProps={{ shrink: true }}
            />
          </Grid>
          <Grid item xs={12} sm={6}>
            <TextField
              fullWidth
              label="Schedule Time"
              type="time"
              value={campaignData.schedule_time}
              onChange={(e) => setCampaignData({ ...campaignData, schedule_time: e.target.value })}
              InputLabelProps={{ shrink: true }}
            />
          </Grid>
        </>
      )}

      <Grid item xs={12} sm={6}>
        <TextField
          fullWidth
          label="Max Messages Per Day"
          type="number"
          value={campaignData.max_messages_per_day}
          onChange={(e) => setCampaignData({ ...campaignData, max_messages_per_day: e.target.value })}
          helperText="Optional rate limiting"
        />
      </Grid>
    </Grid>
  );

  // Step 5: Review
  const renderStep5 = () => (
    <Grid container spacing={3}>
      <Grid item xs={12}>
        <Alert severity="info" sx={{ mb: 2 }}>
          Please review your campaign details before submitting
        </Alert>
      </Grid>

      <Grid item xs={12}>
        <Paper sx={{ p: 2 }}>
          <Typography variant="h6" gutterBottom>Campaign Summary</Typography>
          <Grid container spacing={2}>
            <Grid item xs={6}>
              <Typography variant="body2" color="textSecondary">Campaign Name:</Typography>
              <Typography variant="body1">{campaignData.name}</Typography>
            </Grid>
            <Grid item xs={6}>
              <Typography variant="body2" color="textSecondary">Channel:</Typography>
              <Typography variant="body1">{campaignData.channel}</Typography>
            </Grid>
            <Grid item xs={6}>
              <Typography variant="body2" color="textSecondary">Sender ID:</Typography>
              <Typography variant="body1">{campaignData.sender_id}</Typography>
            </Grid>
            <Grid item xs={6}>
              <Typography variant="body2" color="textSecondary">Priority:</Typography>
              <Chip label={campaignData.priority} size="small" />
            </Grid>
            <Grid item xs={6}>
              <Typography variant="body2" color="textSecondary">Total Recipients:</Typography>
              <Typography variant="h6">{campaignData.total_recipients?.toLocaleString()}</Typography>
            </Grid>
            <Grid item xs={6}>
              <Typography variant="body2" color="textSecondary">Estimated Cost:</Typography>
              <Typography variant="h6">${campaignData.total_recipients * 0.01}</Typography>
            </Grid>
            <Grid item xs={12}>
              <Typography variant="body2" color="textSecondary">Message Preview:</Typography>
              <Paper sx={{ p: 2, mt: 1, bgcolor: 'grey.100' }}>
                <Typography variant="body2">{campaignData.message_content || '(Using template)'}</Typography>
              </Paper>
            </Grid>
            <Grid item xs={6}>
              <Typography variant="body2" color="textSecondary">Schedule:</Typography>
              <Typography variant="body1">
                {campaignData.schedule_type === 'IMMEDIATE' ? 'Send Immediately' : `${campaignData.schedule_date} ${campaignData.schedule_time}`}
              </Typography>
            </Grid>
          </Grid>
        </Paper>
      </Grid>
    </Grid>
  );

  const getStepContent = (step) => {
    switch (step) {
      case 0:
        return renderStep1();
      case 1:
        return renderStep2();
      case 2:
        return renderStep3();
      case 3:
        return renderStep4();
      case 4:
        return renderStep5();
      default:
        return 'Unknown step';
    }
  };

  return (
    <Box>
      <Typography variant="h5" gutterBottom>
        Create New Campaign
      </Typography>
      <Typography variant="body2" color="textSecondary" paragraph>
        Follow the steps below to create and launch a new messaging campaign
      </Typography>

      <Paper sx={{ p: 3, mt: 3 }}>
        <Stepper activeStep={activeStep} sx={{ mb: 4 }}>
          {steps.map((label) => (
            <Step key={label}>
              <StepLabel>{label}</StepLabel>
            </Step>
          ))}
        </Stepper>

        <Box sx={{ minHeight: 400 }}>
          {getStepContent(activeStep)}
        </Box>

        <Box sx={{ display: 'flex', justifyContent: 'space-between', mt: 3 }}>
          <Button
            disabled={activeStep === 0}
            onClick={handleBack}
          >
            Back
          </Button>
          <Box>
            {activeStep === steps.length - 1 ? (
              <Button
                variant="contained"
                onClick={handleSubmit}
                size="large"
              >
                Submit Campaign
              </Button>
            ) : (
              <Button
                variant="contained"
                onClick={handleNext}
              >
                Next
              </Button>
            )}
          </Box>
        </Box>
      </Paper>
    </Box>
  );
}

export default CreateCampaign;
