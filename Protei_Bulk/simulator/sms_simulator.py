#!/usr/bin/env python3
"""
SMS Simulator - Test message sending and delivery simulation
Provides GUI and CLI interfaces for testing the Protei_Bulk platform
"""

import tkinter as tk
from tkinter import ttk, scrolledtext, messagebox
import asyncio
import requests
from datetime import datetime
from typing import Dict, List
import json


class SMSSimulator:
    """SMS Simulator with GUI"""

    def __init__(self, api_url="http://localhost:8080/api/v1"):
        self.api_url = api_url
        self.api_key = ""

        # Create main window
        self.root = tk.Tk()
        self.root.title("Protei_Bulk SMS Simulator")
        self.root.geometry("900x700")

        self._create_widgets()

    def _create_widgets(self):
        """Create GUI widgets"""

        # API Configuration Frame
        config_frame = ttk.LabelFrame(self.root, text="API Configuration", padding=10)
        config_frame.pack(fill="x", padx=10, pady=5)

        ttk.Label(config_frame, text="API URL:").grid(row=0, column=0, sticky="w", padx=5)
        self.api_url_entry = ttk.Entry(config_frame, width=40)
        self.api_url_entry.insert(0, self.api_url)
        self.api_url_entry.grid(row=0, column=1, padx=5)

        ttk.Label(config_frame, text="API Key:").grid(row=1, column=0, sticky="w", padx=5)
        self.api_key_entry = ttk.Entry(config_frame, width=40, show="*")
        self.api_key_entry.grid(row=1, column=1, padx=5)

        ttk.Button(config_frame, text="Test Connection", command=self.test_connection).grid(
            row=0, column=2, rowspan=2, padx=5
        )

        # Message Composition Frame
        msg_frame = ttk.LabelFrame(self.root, text="Message Composition", padding=10)
        msg_frame.pack(fill="both", expand=True, padx=10, pady=5)

        # Left side - Message details
        left_frame = ttk.Frame(msg_frame)
        left_frame.pack(side="left", fill="both", expand=True, padx=5)

        ttk.Label(left_frame, text="From:").grid(row=0, column=0, sticky="w", pady=5)
        self.from_entry = ttk.Entry(left_frame, width=30)
        self.from_entry.insert(0, "1234")
        self.from_entry.grid(row=0, column=1, sticky="ew", pady=5)

        ttk.Label(left_frame, text="To:").grid(row=1, column=0, sticky="w", pady=5)
        self.to_entry = ttk.Entry(left_frame, width=30)
        self.to_entry.insert(0, "9876543210")
        self.to_entry.grid(row=1, column=1, sticky="ew", pady=5)

        ttk.Label(left_frame, text="Message:").grid(row=2, column=0, sticky="nw", pady=5)
        self.message_text = scrolledtext.ScrolledText(left_frame, width=40, height=8)
        self.message_text.insert("1.0", "Hello! This is a test message from Protei_Bulk simulator.")
        self.message_text.grid(row=2, column=1, sticky="ew", pady=5)

        ttk.Label(left_frame, text="Encoding:").grid(row=3, column=0, sticky="w", pady=5)
        self.encoding_var = tk.StringVar(value="GSM7")
        encoding_combo = ttk.Combobox(left_frame, textvariable=self.encoding_var,
                                      values=["GSM7", "UCS2", "ASCII"], state="readonly", width=28)
        encoding_combo.grid(row=3, column=1, sticky="ew", pady=5)

        ttk.Label(left_frame, text="Priority:").grid(row=4, column=0, sticky="w", pady=5)
        self.priority_var = tk.StringVar(value="NORMAL")
        priority_combo = ttk.Combobox(left_frame, textvariable=self.priority_var,
                                      values=["CRITICAL", "HIGH", "NORMAL", "LOW"],
                                      state="readonly", width=28)
        priority_combo.grid(row=4, column=1, sticky="ew", pady=5)

        # Character counter
        self.char_count_label = ttk.Label(left_frame, text="Characters: 0 / Parts: 1")
        self.char_count_label.grid(row=5, column=1, sticky="e", pady=5)
        self.message_text.bind("<KeyRelease>", self.update_char_count)

        # Right side - Handset preview
        right_frame = ttk.LabelFrame(msg_frame, text="Handset Preview", padding=10)
        right_frame.pack(side="right", fill="both", padx=5)

        # Phone display
        phone_canvas = tk.Canvas(right_frame, width=200, height=350, bg="#e0e0e0", relief="sunken", bd=2)
        phone_canvas.pack(pady=10)

        # Screen area
        phone_canvas.create_rectangle(10, 40, 190, 300, fill="white", outline="black")
        phone_canvas.create_text(100, 20, text="SMS Preview", font=("Arial", 10, "bold"))

        # Message preview
        self.preview_text_id = phone_canvas.create_text(
            100, 150, text="Message preview will appear here",
            font=("Arial", 9), width=160, justify="left"
        )
        self.preview_canvas = phone_canvas

        # Update preview
        self.message_text.bind("<KeyRelease>", self.update_preview)

        # Action Buttons Frame
        action_frame = ttk.Frame(self.root)
        action_frame.pack(fill="x", padx=10, pady=5)

        ttk.Button(action_frame, text="Send Message", command=self.send_message,
                  style="Accent.TButton").pack(side="left", padx=5)
        ttk.Button(action_frame, text="Send Bulk (CSV)", command=self.send_bulk).pack(side="left", padx=5)
        ttk.Button(action_frame, text="Clear", command=self.clear_form).pack(side="left", padx=5)

        # Response/Log Frame
        log_frame = ttk.LabelFrame(self.root, text="Response Log", padding=10)
        log_frame.pack(fill="both", expand=True, padx=10, pady=5)

        self.log_text = scrolledtext.ScrolledText(log_frame, height=10)
        self.log_text.pack(fill="both", expand=True)

        # Status Bar
        self.status_bar = ttk.Label(self.root, text="Ready", relief="sunken", anchor="w")
        self.status_bar.pack(fill="x", side="bottom")

    def update_char_count(self, event=None):
        """Update character count and parts"""
        text = self.message_text.get("1.0", "end-1c")
        length = len(text)

        # Calculate SMS parts (160 chars for GSM7, 70 for UCS2)
        encoding = self.encoding_var.get()
        max_chars = 160 if encoding == "GSM7" else 70
        parts = max(1, (length + max_chars - 1) // max_chars)

        self.char_count_label.config(text=f"Characters: {length} / Parts: {parts}")

    def update_preview(self, event=None):
        """Update handset preview"""
        text = self.message_text.get("1.0", "end-1c")
        from_num = self.from_entry.get()

        preview = f"From: {from_num}\n\n{text}"
        self.preview_canvas.itemconfig(self.preview_text_id, text=preview)

    def log(self, message: str, level: str = "INFO"):
        """Add message to log"""
        timestamp = datetime.now().strftime("%H:%M:%S")
        log_entry = f"[{timestamp}] [{level}] {message}\n"

        self.log_text.insert("end", log_entry)
        self.log_text.see("end")

        # Color coding
        if level == "ERROR":
            self.log_text.tag_add("error", "end-2l", "end-1l")
            self.log_text.tag_config("error", foreground="red")
        elif level == "SUCCESS":
            self.log_text.tag_add("success", "end-2l", "end-1l")
            self.log_text.tag_config("success", foreground="green")

    def test_connection(self):
        """Test API connection"""
        self.status_bar.config(text="Testing connection...")
        self.log("Testing connection to API...")

        try:
            url = f"{self.api_url_entry.get()}/health"
            response = requests.get(url, timeout=5)

            if response.status_code == 200:
                data = response.json()
                self.log(f"Connection successful! Version: {data.get('version', 'unknown')}", "SUCCESS")
                self.status_bar.config(text="Connected")
                messagebox.showinfo("Success", "API connection successful!")
            else:
                self.log(f"Connection failed: HTTP {response.status_code}", "ERROR")
                self.status_bar.config(text="Connection failed")

        except Exception as e:
            self.log(f"Connection error: {str(e)}", "ERROR")
            self.status_bar.config(text="Connection error")
            messagebox.showerror("Error", f"Connection failed: {str(e)}")

    def send_message(self):
        """Send a single message"""
        # Validate inputs
        if not self.from_entry.get() or not self.to_entry.get():
            messagebox.showerror("Error", "From and To fields are required")
            return

        message_text = self.message_text.get("1.0", "end-1c")
        if not message_text:
            messagebox.showerror("Error", "Message text is required")
            return

        # Prepare request
        data = {
            "from": self.from_entry.get(),
            "to": self.to_entry.get(),
            "text": message_text,
            "encoding": self.encoding_var.get(),
            "priority": self.priority_var.get()
        }

        headers = {}
        if self.api_key_entry.get():
            headers["X-API-Key"] = self.api_key_entry.get()

        self.log(f"Sending message to {data['to']}...")
        self.status_bar.config(text="Sending message...")

        try:
            url = f"{self.api_url_entry.get()}/messages"
            response = requests.post(url, json=data, headers=headers, timeout=10)

            if response.status_code in [200, 201]:
                result = response.json()
                message_id = result.get("message_id", "unknown")
                self.log(f"Message sent successfully! ID: {message_id}", "SUCCESS")
                self.log(f"Response: {json.dumps(result, indent=2)}")
                self.status_bar.config(text="Message sent successfully")
                messagebox.showinfo("Success", f"Message sent!\nMessage ID: {message_id}")
            else:
                self.log(f"Send failed: HTTP {response.status_code}", "ERROR")
                self.log(f"Response: {response.text}", "ERROR")
                self.status_bar.config(text="Send failed")
                messagebox.showerror("Error", f"Send failed: {response.text}")

        except Exception as e:
            self.log(f"Send error: {str(e)}", "ERROR")
            self.status_bar.config(text="Send error")
            messagebox.showerror("Error", f"Send failed: {str(e)}")

    def send_bulk(self):
        """Send bulk messages from CSV"""
        from tkinter import filedialog

        filename = filedialog.askopenfilename(
            title="Select CSV file",
            filetypes=[("CSV files", "*.csv"), ("All files", "*.*")]
        )

        if not filename:
            return

        self.log(f"Loading bulk messages from: {filename}")
        # Implementation would parse CSV and send messages
        messagebox.showinfo("Info", "Bulk send feature - parse CSV and send messages (to be implemented)")

    def clear_form(self):
        """Clear the form"""
        self.message_text.delete("1.0", "end")
        self.to_entry.delete(0, "end")
        self.update_char_count()
        self.update_preview()

    def run(self):
        """Run the simulator"""
        self.root.mainloop()


class CLI_SMSSimulator:
    """Command-line SMS Simulator"""

    def __init__(self, api_url="http://localhost:8080/api/v1", api_key=None):
        self.api_url = api_url
        self.api_key = api_key

    def send_message(self, from_addr: str, to_addr: str, text: str, **kwargs):
        """Send a single message"""
        data = {
            "from": from_addr,
            "to": to_addr,
            "text": text,
            **kwargs
        }

        headers = {}
        if self.api_key:
            headers["X-API-Key"] = self.api_key

        try:
            response = requests.post(f"{self.api_url}/messages", json=data, headers=headers)
            response.raise_for_status()
            return response.json()
        except Exception as e:
            print(f"Error: {e}")
            return None

    def send_bulk(self, messages: List[Dict]):
        """Send multiple messages"""
        results = []
        for msg in messages:
            result = self.send_message(**msg)
            results.append(result)
        return results


if __name__ == "__main__":
    import sys

    if len(sys.argv) > 1 and sys.argv[1] == "--cli":
        # CLI mode
        simulator = CLI_SMSSimulator()
        result = simulator.send_message(
            from_addr="1234",
            to_addr="9876543210",
            text="Test message from CLI simulator"
        )
        print(f"Result: {result}")
    else:
        # GUI mode
        simulator = SMSSimulator()
        simulator.run()
