// JavaScript for Secure Packager Demo

// API base URL
const API_BASE = '/api';

// Utility functions
function showStatus(elementId, message, type = 'info') {
    const element = document.getElementById(elementId);
    if (!element) return;
    
    element.innerHTML = `<div class="alert alert-${type} fade-in">${message}</div>`;
}

function showLoading(elementId) {
    const element = document.getElementById(elementId);
    if (!element) return;
    
    element.innerHTML = '<div class="text-center"><div class="loading"></div> Processing...</div>';
}

function formatFileSize(bytes) {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function formatFileType(type) {
    return type === 'directory' ? 'folder' : 'file';
}

// API functions
async function apiCall(endpoint, method = 'GET', data = null) {
    const options = {
        method: method,
        headers: {
            'Content-Type': 'application/json',
        },
    };
    
    if (data) {
        options.body = JSON.stringify(data);
    }
    
    try {
        const response = await fetch(`${API_BASE}${endpoint}`, options);
        const result = await response.json();
        
        if (!response.ok) {
            throw new Error(result.message || `HTTP ${response.status}`);
        }
        
        return result;
    } catch (error) {
        console.error('API call failed:', error);
        throw error;
    }
}

// Key generation
async function generateKeys() {
    const keySize = parseInt(document.getElementById('keySize').value);
    showLoading('keyStatus');
    
    try {
        const result = await apiCall('/keys/generate', 'POST', { key_size: keySize });
        showStatus('keyStatus', `✅ ${result.message}`, 'success');
        
        // Update key files display
        const keyFiles = document.getElementById('keyFiles');
        keyFiles.innerHTML = `
            <div class="file-list">
                <div class="file-item">
                    <span class="file-name">customer_private.pem</span>
                    <span class="file-type">private key</span>
                </div>
                <div class="file-item">
                    <span class="file-name">customer_public.pem</span>
                    <span class="file-type">public key</span>
                </div>
                <div class="file-item">
                    <span class="file-name">vendor_private.pem</span>
                    <span class="file-type">private key</span>
                </div>
                <div class="file-item">
                    <span class="file-name">vendor_public.pem</span>
                    <span class="file-type">public key</span>
                </div>
            </div>
        `;
    } catch (error) {
        // Show informative message about pre-generated keys
        showStatus('keyStatus', `ℹ️ ${error.message}`, 'info');
        
        // Update key files display to show available keys
        const keyFiles = document.getElementById('keyFiles');
        keyFiles.innerHTML = `
            <div class="file-list">
                <div class="file-item">
                    <span class="file-name">customer_private.pem</span>
                    <span class="file-type">private key</span>
                </div>
                <div class="file-item">
                    <span class="file-name">customer_public.pem</span>
                    <span class="file-type">public key</span>
                </div>
                <div class="file-item">
                    <span class="file-name">vendor_private.pem</span>
                    <span class="file-type">private key</span>
                </div>
                <div class="file-item">
                    <span class="file-name">vendor_public.pem</span>
                    <span class="file-type">public key</span>
                </div>
            </div>
        `;
    }
}

// Upload files
async function uploadFiles() {
    const fileInput = document.getElementById('fileUpload');
    const files = fileInput.files;
    
    if (files.length === 0) {
        showStatus('fileStatus', '❌ Please select files to upload', 'danger');
        return;
    }
    
    showLoading('fileStatus');
    
    try {
        const formData = new FormData();
        for (let i = 0; i < files.length; i++) {
            formData.append('files', files[i]);
        }
        
        const result = await fetch('/api/files/upload', {
            method: 'POST',
            body: formData
        });
        
        const response = await result.json();
        
        if (response.success) {
            showStatus('fileStatus', `✅ ${response.message}`, 'success');
            
            // Update created files display
            const createdFiles = document.getElementById('createdFiles');
            const fileList = response.data.map(file => `<li>${file}</li>`).join('');
            createdFiles.innerHTML = `
                <div class="alert alert-success">
                    <h6><i class="fas fa-check-circle me-2"></i>Files Uploaded</h6>
                    <ul class="mb-0">${fileList}</ul>
                </div>
            `;
            
            // Clear file input
            fileInput.value = '';
        } else {
            showStatus('fileStatus', `❌ ${response.message}`, 'danger');
        }
    } catch (error) {
        showStatus('fileStatus', `❌ Upload failed: ${error.message}`, 'danger');
    }
}

async function clearDataDirectory() {
    if (!confirm('Are you sure you want to clear all files in the data directory? This action cannot be undone.')) {
        return;
    }
    
    showLoading('fileStatus');
    
    try {
        const result = await fetch('/api/files/clear-data', {
            method: 'POST'
        });
        
        const response = await result.json();
        
        if (response.success) {
            showStatus('fileStatus', `✅ ${response.message}`, 'success');
            
            // Clear created files display
            const createdFiles = document.getElementById('createdFiles');
            createdFiles.innerHTML = '';
        } else {
            showStatus('fileStatus', `❌ ${response.message}`, 'danger');
        }
    } catch (error) {
        showStatus('fileStatus', `❌ Clear failed: ${error.message}`, 'danger');
    }
}

// Create files
async function createFiles() {
    const content = document.getElementById('fileContent').value;
    showLoading('fileStatus');
    
    try {
        const result = await apiCall('/files/create', 'POST', { content: content });
        showStatus('fileStatus', `✅ ${result.message}`, 'success');
        
        // Show created files
        const createdFiles = document.getElementById('createdFiles');
        createdFiles.innerHTML = `
            <div class="alert alert-info">
                <h6><i class="fas fa-files me-2"></i>Created Files</h6>
                <div class="file-list">
                    <div class="file-item">
                        <span class="file-name">sample.txt</span>
                        <span class="file-type">text file</span>
                    </div>
                    <div class="file-item">
                        <span class="file-name">config.json</span>
                        <span class="file-type">json file</span>
                    </div>
                </div>
            </div>
        `;
    } catch (error) {
        showStatus('fileStatus', `❌ ${error.message}`, 'danger');
    }
}

// Package files
async function packageFiles() {
    const useLicensing = document.getElementById('useLicensing').checked;
    showLoading('packageStatus');
    
    try {
        const result = await apiCall('/package', 'POST', { use_licensing: useLicensing });
        showStatus('packageStatus', `✅ ${result.message}`, 'success');
        
        // Show packaged files
        const packagedFiles = document.getElementById('packagedFiles');
        packagedFiles.innerHTML = `
            <div class="alert alert-info">
                <h6><i class="fas fa-archive me-2"></i>Packaged Files</h6>
                <div class="file-list">
                    <div class="file-item">
                        <span class="file-name">encrypted_files.zip</span>
                        <span class="file-type">encrypted archive</span>
                    </div>
                    <div class="file-item">
                        <span class="file-name">wrapped_key.bin</span>
                        <span class="file-type">encrypted key</span>
                    </div>
                    ${useLicensing ? `
                    <div class="file-item">
                        <span class="file-name">manifest.json</span>
                        <span class="file-type">license manifest</span>
                    </div>
                    <div class="file-item">
                        <span class="file-name">vendor_public.pem</span>
                        <span class="file-type">vendor public key</span>
                    </div>
                    ` : ''}
                </div>
            </div>
        `;
    } catch (error) {
        showStatus('packageStatus', `❌ ${error.message}`, 'danger');
    }
}

// Issue token
async function issueToken() {
    const company = document.getElementById('company').value;
    const email = document.getElementById('email').value;
    const expiryDays = parseInt(document.getElementById('expiryDays').value);
    
    if (!company || !email) {
        showStatus('tokenStatus', '❌ Please fill in all required fields', 'danger');
        return;
    }
    
    showLoading('tokenStatus');
    
    try {
        const result = await apiCall('/token/issue', 'POST', {
            company: company,
            email: email,
            expiry_days: expiryDays
        });
        showStatus('tokenStatus', `✅ ${result.message}`, 'success');
    } catch (error) {
        showStatus('tokenStatus', `❌ ${error.message}`, 'danger');
    }
}

// Unpack files
async function unpackFiles() {
    const useLicensing = document.getElementById('unpackLicensing').checked;
    showLoading('unpackStatus');
    
    try {
        const result = await apiCall('/unpack', 'POST', { use_licensing: useLicensing });
        showStatus('unpackStatus', `✅ ${result.message}`, 'success');
        
        // Show actual unpacked files by refreshing the file list
        displayUnpackedFiles();
    } catch (error) {
        showStatus('unpackStatus', `❌ ${error.message}`, 'danger');
    }
}

// Display unpacked files
async function displayUnpackedFiles() {
    try {
        const result = await apiCall('/files/decrypted');
        const files = result.data;
        
        const unpackedFiles = document.getElementById('unpackedFiles');
        
        if (!files || files.length === 0) {
            unpackedFiles.innerHTML = '<div class="alert alert-info">No files found in decrypted directory</div>';
            return;
        }
        
        const fileListHTML = files.map(file => `
            <div class="file-item d-flex justify-content-between align-items-center p-2 border rounded mb-2">
                <div>
                    <i class="fas fa-${formatFileType(file.type)} me-2"></i>
                    <strong>${file.name}</strong>
                    <small class="text-muted ms-2">${formatFileSize(file.size)}</small>
                </div>
                <div class="btn-group" role="group">
                    <button class="btn btn-sm btn-outline-primary" onclick="selectFile('${file.name}')">
                        <i class="fas fa-eye me-1"></i>View
                    </button>
                    <button class="btn btn-sm btn-outline-success" onclick="downloadDecryptedFile('${file.name}')">
                        <i class="fas fa-download me-1"></i>Download
                    </button>
                </div>
            </div>
        `).join('');
        
        unpackedFiles.innerHTML = `
            <div class="alert alert-success">
                <h6><i class="fas fa-check-circle me-2"></i>Unpacked Files</h6>
                <div class="file-list">${fileListHTML}</div>
            </div>
        `;
    } catch (error) {
        const unpackedFiles = document.getElementById('unpackedFiles');
        unpackedFiles.innerHTML = `<div class="alert alert-danger">❌ Failed to list unpacked files: ${error.message}</div>`;
    }
}

// Clear output files
async function clearOutputFiles() {
    if (!confirm('Are you sure you want to clear all output files (encrypted zips and decrypted files)? This action cannot be undone.')) {
        return;
    }
    
    showLoading('packageStatus');
    
    try {
        const result = await fetch('/api/files/clear-output', {
            method: 'POST'
        });
        
        const response = await result.json();
        
        if (response.success) {
            showStatus('packageStatus', `✅ ${response.message}`, 'success');
            
            // Clear packaged files display
            const packagedFiles = document.getElementById('packagedFiles');
            packagedFiles.innerHTML = '';
        } else {
            showStatus('packageStatus', `❌ ${response.message}`, 'danger');
        }
    } catch (error) {
        showStatus('packageStatus', `❌ Clear failed: ${error.message}`, 'danger');
    }
}

// Clear decrypted files
async function clearDecryptedFiles() {
    if (!confirm('Are you sure you want to clear all decrypted files? This action cannot be undone.')) {
        return;
    }
    
    showLoading('unpackStatus');
    
    try {
        const result = await fetch('/api/files/clear-decrypted', {
            method: 'POST'
        });
        
        const response = await result.json();
        
        if (response.success) {
            showStatus('unpackStatus', `✅ ${response.message}`, 'success');
            
            // Clear unpacked files display
            const unpackedFiles = document.getElementById('unpackedFiles');
            unpackedFiles.innerHTML = '';
        } else {
            showStatus('unpackStatus', `❌ ${response.message}`, 'danger');
        }
    } catch (error) {
        showStatus('unpackStatus', `❌ Clear failed: ${error.message}`, 'danger');
    }
}

// File browser functions
async function refreshFiles() {
    const directory = document.getElementById('directory').value;
    showLoading('fileList');
    
    try {
        const result = await apiCall(`/files/${directory}`);
        displayFileList(result.data);
    } catch (error) {
        showStatus('fileList', `❌ ${error.message}`, 'danger');
    }
}

function displayFileList(files) {
    const fileList = document.getElementById('fileList');
    const directory = document.getElementById('directory').value;
    
    if (!files || files.length === 0) {
        fileList.innerHTML = '<div class="alert alert-info">No files found in this directory</div>';
        return;
    }
    
    const fileListHTML = files.map(file => {
        const isDownloadable = directory === 'output' && (file.name.endsWith('.zip') || file.name.endsWith('.encrypted'));
        return `
            <div class="file-item d-flex justify-content-between align-items-center p-2 border rounded mb-2">
                <div>
                    <i class="fas fa-${formatFileType(file.type)} me-2"></i>
                    <strong>${file.name}</strong>
                    <small class="text-muted ms-2">${formatFileSize(file.size)}</small>
                </div>
                <div class="btn-group" role="group">
                    <button class="btn btn-sm btn-outline-primary" onclick="selectFile('${file.name}')">
                        <i class="fas fa-eye me-1"></i>View
                    </button>
                    ${isDownloadable ? `
                    <button class="btn btn-sm btn-outline-success" onclick="downloadFileDirect('${file.name}')">
                        <i class="fas fa-download me-1"></i>Download
                    </button>
                    ` : ''}
                </div>
            </div>
        `;
    }).join('');
    
    fileList.innerHTML = `
        <div class="alert alert-info">
            <h6><i class="fas fa-folder me-2"></i>Files in ${directory} directory</h6>
            <div class="file-list">${fileListHTML}</div>
        </div>
    `;
}

async function readFile() {
    const filename = document.getElementById('filename').value;
    const directory = document.getElementById('directory').value;
    
    console.log('readFile called with filename:', filename, 'directory:', directory);
    
    if (!filename) {
        showStatus('fileContent', '❌ Please enter a filename', 'danger');
        return;
    }
    
    showLoading('fileContent');
    
    try {
        const result = await apiCall('/files/read', 'POST', {
            filename: filename,
            directory: directory
        });
        
        const fileContent = document.getElementById('fileContent');
        fileContent.innerHTML = `
            <div class="alert alert-info">
                <h6><i class="fas fa-file me-2"></i>Content of ${filename}</h6>
                <div class="file-content">${result.data}</div>
            </div>
        `;
        
        // Show download button for output directory files (encrypted files)
        const downloadBtn = document.getElementById('downloadBtn');
        if (directory === 'output' && (filename.endsWith('.zip') || filename.endsWith('.encrypted'))) {
            downloadBtn.style.display = 'inline-block';
        } else {
            downloadBtn.style.display = 'none';
        }
    } catch (error) {
        showStatus('fileContent', `❌ ${error.message}`, 'danger');
    }
}

async function downloadFile() {
    const filename = document.getElementById('filename').value;
    const directory = document.getElementById('directory').value;
    
    if (!filename) {
        showStatus('fileContent', '❌ Please enter a filename', 'danger');
        return;
    }
    
    if (directory !== 'output') {
        showStatus('fileContent', '❌ Download is only available for files in the output directory', 'danger');
        return;
    }
    
    try {
        // Create download link
        const downloadUrl = `/api/files/download/${encodeURIComponent(filename)}`;
        const link = document.createElement('a');
        link.href = downloadUrl;
        link.download = filename;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        
        showStatus('fileContent', `✅ Download started for ${filename}`, 'success');
    } catch (error) {
        showStatus('fileContent', `❌ Download failed: ${error.message}`, 'danger');
    }
}

async function downloadFileDirect(filename) {
    const directory = document.getElementById('directory').value;
    
    if (directory !== 'output') {
        showStatus('fileContent', '❌ Download is only available for files in the output directory', 'danger');
        return;
    }
    
    try {
        // Create download link
        const downloadUrl = `/api/files/download/${encodeURIComponent(filename)}`;
        const link = document.createElement('a');
        link.href = downloadUrl;
        link.download = filename;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        
        showStatus('fileContent', `✅ Download started for ${filename}`, 'success');
    } catch (error) {
        showStatus('fileContent', `❌ Download failed: ${error.message}`, 'danger');
    }
}

function selectFile(filename) {
    console.log('selectFile called with:', filename);
    // Set the filename in the input field and trigger file reading
    document.getElementById('filename').value = filename;
    console.log('Filename set to:', document.getElementById('filename').value);
    readFile();
}

// Upload and unpack functionality
async function uploadAndUnpack() {
    const encryptedZip = document.getElementById('encryptedZipUpload').files[0];
    const customerPrivate = document.getElementById('customerPrivateUpload').files[0];
    const vendorPublic = document.getElementById('vendorPublicUpload').files[0];
    const token = document.getElementById('tokenUpload').files[0];
    const useLicensing = document.getElementById('useLicensingUpload').checked;
    
    // Validate required files
    if (!encryptedZip) {
        showStatus('uploadUnpackStatus', '❌ Please select an encrypted ZIP file', 'danger');
        return;
    }
    if (!customerPrivate) {
        showStatus('uploadUnpackStatus', '❌ Please select a customer private key', 'danger');
        return;
    }
    if (useLicensing) {
        if (!vendorPublic) {
            showStatus('uploadUnpackStatus', '❌ Please select a vendor public key for licensing', 'danger');
            return;
        }
        if (!token) {
            showStatus('uploadUnpackStatus', '❌ Please select a license token for licensing', 'danger');
            return;
        }
    }
    
    showLoading('uploadUnpackStatus');
    
    try {
        const formData = new FormData();
        formData.append('encryptedZip', encryptedZip);
        formData.append('customerPrivate', customerPrivate);
        if (useLicensing) {
            formData.append('vendorPublic', vendorPublic);
            formData.append('token', token);
        }
        formData.append('useLicensing', useLicensing);
        
        const result = await fetch('/api/files/upload-unpack', {
            method: 'POST',
            body: formData
        });
        
        const response = await result.json();
        
        if (response.success) {
            showStatus('uploadUnpackStatus', `✅ ${response.message}`, 'success');
            
            // Show unpacked files
            displayUnpackedFiles();
        } else {
            showStatus('uploadUnpackStatus', `❌ ${response.message}`, 'danger');
        }
    } catch (error) {
        showStatus('uploadUnpackStatus', `❌ Upload and unpack failed: ${error.message}`, 'danger');
    }
}

function clearUploads() {
    document.getElementById('encryptedZipUpload').value = '';
    document.getElementById('customerPrivateUpload').value = '';
    document.getElementById('vendorPublicUpload').value = '';
    document.getElementById('tokenUpload').value = '';
    document.getElementById('uploadUnpackStatus').innerHTML = '';
    document.getElementById('unpackedFilesList').innerHTML = '';
}

async function displayUnpackedFiles() {
    try {
        const result = await apiCall('/files/decrypted');
        const files = result.data;
        
        if (!files || files.length === 0) {
            document.getElementById('unpackedFilesList').innerHTML = '<div class="alert alert-info">No files found in decrypted directory</div>';
            return;
        }
        
        const fileListHTML = files.map(file => `
            <div class="file-item d-flex justify-content-between align-items-center p-2 border rounded mb-2">
                <div>
                    <i class="fas fa-${formatFileType(file.type)} me-2"></i>
                    <strong>${file.name}</strong>
                    <small class="text-muted ms-2">${formatFileSize(file.size)}</small>
                </div>
                <div class="btn-group" role="group">
                    <button class="btn btn-sm btn-outline-success" onclick="downloadDecryptedFile('${file.name}')">
                        <i class="fas fa-download me-1"></i>Download
                    </button>
                </div>
            </div>
        `).join('');
        
        document.getElementById('unpackedFilesList').innerHTML = `
            <div class="alert alert-success">
                <h6><i class="fas fa-check-circle me-2"></i>Unpacked Files</h6>
                <div class="file-list">${fileListHTML}</div>
            </div>
        `;
    } catch (error) {
        document.getElementById('unpackedFilesList').innerHTML = `<div class="alert alert-danger">❌ Failed to list unpacked files: ${error.message}</div>`;
    }
}

async function downloadDecryptedFile(filename) {
    try {
        const downloadUrl = `/api/files/download/${encodeURIComponent(filename)}?dir=decrypted`;
        const link = document.createElement('a');
        link.href = downloadUrl;
        link.download = filename;
        document.body.appendChild(link);
        link.click();
        document.body.removeChild(link);
        
        showStatus('uploadUnpackStatus', `✅ Download started for ${filename}`, 'success');
    } catch (error) {
        showStatus('uploadUnpackStatus', `❌ Download failed: ${error.message}`, 'danger');
    }
}

// Complete workflow
async function runCompleteWorkflow() {
    const workflowOutput = document.getElementById('workflowOutput');
    workflowOutput.innerHTML = '<div class="text-center"><div class="loading"></div> Running complete workflow...</div>';
    
    try {
        const result = await apiCall('/workflow/complete', 'POST');
        
        workflowOutput.innerHTML = `
            <div class="alert alert-success">
                <h6><i class="fas fa-check-circle me-2"></i>Workflow Complete</h6>
                <div class="workflow-output">${result.message}</div>
            </div>
        `;
    } catch (error) {
        workflowOutput.innerHTML = `
            <div class="alert alert-danger">
                <h6><i class="fas fa-exclamation-triangle me-2"></i>Workflow Failed</h6>
                <div class="workflow-output">${error.message}</div>
            </div>
        `;
    }
}

// Event listeners
document.addEventListener('DOMContentLoaded', function() {
    // Key size slider
    const keySizeSlider = document.getElementById('keySize');
    const keySizeValue = document.getElementById('keySizeValue');
    keySizeSlider.addEventListener('input', function() {
        keySizeValue.textContent = this.value;
    });
    
    // Expiry days slider
    const expiryDaysSlider = document.getElementById('expiryDays');
    const expiryDaysValue = document.getElementById('expiryDaysValue');
    expiryDaysSlider.addEventListener('input', function() {
        expiryDaysValue.textContent = this.value;
    });
    
    // Initialize file browser
    refreshFiles();
    
    // Add smooth scrolling for anchor links
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function (e) {
            e.preventDefault();
            const target = document.querySelector(this.getAttribute('href'));
            if (target) {
                target.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });
            }
        });
    });
    
    // Add keyboard shortcuts
    document.addEventListener('keydown', function(e) {
        // Ctrl+Enter to run current tab's main action
        if (e.ctrlKey && e.key === 'Enter') {
            const activeTab = document.querySelector('.nav-link.active');
            if (activeTab) {
                const tabId = activeTab.getAttribute('data-bs-target').replace('#', '');
                switch (tabId) {
                    case 'keys':
                        generateKeys();
                        break;
                    case 'files':
                        createFiles();
                        break;
                    case 'package':
                        packageFiles();
                        break;
                    case 'token':
                        issueToken();
                        break;
                    case 'unpack':
                        unpackFiles();
                        break;
                    case 'workflow':
                        runCompleteWorkflow();
                        break;
                }
            }
        }
    });
    
    // Add tooltips
    const tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'));
    tooltipTriggerList.map(function (tooltipTriggerEl) {
        return new bootstrap.Tooltip(tooltipTriggerEl);
    });
});

// Utility function to copy text to clipboard
function copyToClipboard(text) {
    navigator.clipboard.writeText(text).then(function() {
        // Show success message
        const toast = document.createElement('div');
        toast.className = 'toast position-fixed top-0 end-0 m-3';
        toast.innerHTML = `
            <div class="toast-header">
                <i class="fas fa-check-circle text-success me-2"></i>
                <strong class="me-auto">Copied!</strong>
                <button type="button" class="btn-close" data-bs-dismiss="toast"></button>
            </div>
            <div class="toast-body">
                Text copied to clipboard
            </div>
        `;
        document.body.appendChild(toast);
        const bsToast = new bootstrap.Toast(toast);
        bsToast.show();
        
        // Remove toast after it's hidden
        toast.addEventListener('hidden.bs.toast', function() {
            document.body.removeChild(toast);
        });
    });
}

// Add copy buttons to code blocks
document.addEventListener('DOMContentLoaded', function() {
    const codeBlocks = document.querySelectorAll('code, pre');
    codeBlocks.forEach(function(block) {
        if (block.textContent.length > 20) { // Only add copy button for longer code blocks
            const copyButton = document.createElement('button');
            copyButton.className = 'btn btn-sm btn-outline-secondary position-absolute top-0 end-0 m-2';
            copyButton.innerHTML = '<i class="fas fa-copy"></i>';
            copyButton.title = 'Copy to clipboard';
            copyButton.onclick = function() {
                copyToClipboard(block.textContent);
            };
            
            const container = document.createElement('div');
            container.className = 'position-relative';
            container.appendChild(block.cloneNode(true));
            container.appendChild(copyButton);
            
            block.parentNode.replaceChild(container, block);
        }
    });
});
