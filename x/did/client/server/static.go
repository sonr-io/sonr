package server

// webAuthnRegistrationHTML contains the HTML template for WebAuthn registration
const webAuthnRegistrationHTML = `<!DOCTYPE html>
<html class="dark">
<head>
    <title>Sonr Local Registration</title>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <script src="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"></script>
    <style>
        :root {
            --sonr-primary: #17c2ff;
            --sonr-primary-hover: #0ea5e9;
            --sonr-primary-glow: rgba(23, 194, 255, 0.3);
        }
        
        body {
            background: linear-gradient(135deg, #0f172a 0%, #1e293b 100%);
        }
        
        .glow {
            box-shadow: 0 0 20px var(--sonr-primary-glow);
        }
        
        .pulse-primary {
            animation: pulse-primary 2s ease-in-out infinite;
        }
        
        @keyframes pulse-primary {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.6; }
        }
    </style>
</head>
<body class="min-h-screen flex items-center justify-center bg-slate-900 text-white font-sans">
    <div class="bg-slate-800 rounded-xl p-8 shadow-2xl border border-slate-700 max-w-md w-full mx-4 glow">
        <div class="text-center space-y-6">
            <!-- Header -->
            <div class="space-y-2">
                <h1 class="text-3xl font-bold text-white">Sonr Registration</h1>
                <div class="h-1 bg-gradient-to-r from-[#17c2ff] to-[#0ea5e9] rounded-full mx-auto w-24"></div>
            </div>
            
            <!-- User Info -->
            <div class="bg-slate-700 rounded-lg p-4 border border-slate-600">
                <p class="text-slate-300 text-sm font-medium mb-1">Registering User</p>
                <p id="username-display" class="text-[#17c2ff] text-xl font-bold">{{.Username}}</p>
            </div>
            
            <!-- Status Section -->
            <div class="space-y-4">
                <div id="status" class="text-[#17c2ff] font-semibold text-lg pulse-primary">
                    Initializing WebAuthn registration...
                </div>
                
                <div id="instructions" class="text-slate-300 text-sm leading-relaxed">
                    Please follow your browser and authenticator prompts.
                </div>
                
                <!-- Progress Indicator -->
                <div class="w-full bg-slate-700 rounded-full h-2">
                    <div id="progress" class="bg-gradient-to-r from-[#17c2ff] to-[#0ea5e9] h-2 rounded-full w-0 transition-all duration-500"></div>
                </div>
                
                <!-- Timeout Display -->
                <div id="timeout" class="text-slate-400 text-xs font-mono"></div>
            </div>
        </div>
    </div>

    <!-- Load SimpleWebAuthn for WebAuthn operations -->
    <script src="https://unpkg.com/@simplewebauthn/browser@9.0.1/dist/bundle/index.umd.min.js"></script>
    <!-- Load @sonr.io/es for presets and utilities -->
    <script type="module" src="https://unpkg.com/@sonr.io/es@latest/dist/autoloader.js"></script>
    
    <script>
        // Support both username and identifier parameters
        const urlParams = new URLSearchParams(window.location.search);
        const username = urlParams.get('identifier') || urlParams.get('username') || '{{.Username}}';
        const rpId = '{{.RPID}}';
        const API_URL = window.location.origin; // Use current origin as API URL
        
        // Update the username display if it's from URL params
        if (urlParams.get('identifier') || urlParams.get('username')) {
            document.getElementById('username-display').textContent = username;
        }
        
        function updateStatus(message, type = 'info') {
            const statusEl = document.getElementById('status');
            const progressEl = document.getElementById('progress');
            
            statusEl.textContent = message;
            
            // Remove all existing classes and add new ones based on type
            statusEl.className = 'font-semibold text-lg';
            
            switch(type) {
                case 'success':
                    statusEl.className += ' text-green-400';
                    statusEl.classList.remove('pulse-primary');
                    progressEl.style.width = '100%';
                    progressEl.className = 'bg-gradient-to-r from-green-400 to-green-500 h-2 rounded-full transition-all duration-500';
                    break;
                case 'error':
                    statusEl.className += ' text-red-400';
                    statusEl.classList.remove('pulse-primary');
                    progressEl.style.width = '100%';
                    progressEl.className = 'bg-gradient-to-r from-red-400 to-red-500 h-2 rounded-full transition-all duration-500';
                    break;
                case 'processing':
                    statusEl.className += ' text-[#17c2ff] pulse-primary';
                    progressEl.style.width = '75%';
                    break;
                default: // info
                    statusEl.className += ' text-[#17c2ff] pulse-primary';
                    progressEl.style.width = '25%';
            }
        }
        
        function updateInstructions(message) {
            const instructionsEl = document.getElementById('instructions');
            instructionsEl.textContent = message;
            instructionsEl.className = 'text-slate-300 text-sm leading-relaxed';
        }
        
        function updateTimeout(seconds) {
            const timeoutEl = document.getElementById('timeout');
            if (seconds > 0) {
                timeoutEl.textContent = 'Timeout in ' + seconds + 's';
                timeoutEl.className = 'text-slate-400 text-xs font-mono';
            } else {
                timeoutEl.textContent = 'Registration timed out';
                timeoutEl.className = 'text-red-400 text-xs font-mono';
            }
        }
        
        // Start countdown timer
        let timeoutSeconds = 30;
        const countdownInterval = setInterval(() => {
            updateTimeout(timeoutSeconds);
            timeoutSeconds--;
            if (timeoutSeconds < 0) {
                clearInterval(countdownInterval);
                updateStatus('Registration timed out', 'error');
                updateInstructions('Please return to the CLI and try again.');
            }
        }, 1000);

        async function startRegistration() {
            try {
                // Check if SimpleWebAuthn is loaded
                if (!window.SimpleWebAuthnBrowser) {
                    throw new Error('Failed to load WebAuthn library');
                }
                
                // Check WebAuthn support
                const isSupported = window.SimpleWebAuthnBrowser.browserSupportsWebAuthn();
                if (!isSupported) {
                    throw new Error('WebAuthn is not supported in this browser. Please use a modern browser like Chrome, Firefox, Safari, or Edge.');
                }
                
                // Check if platform authenticator is available
                const isAvailable = await window.SimpleWebAuthnBrowser.platformAuthenticatorIsAvailable();
                if (!isAvailable) {
                    updateStatus('Platform authenticator not available', 'info');
                    updateInstructions('You can use a security key or your phone via QR code to create a passkey.');
                }
                
                updateStatus('Initializing passkey registration...', 'info');
                updateInstructions('Preparing your authentication request...');
                
                // Hybrid approach: Use Sonr presets but local server endpoints
                updateStatus('Initializing passkey registration...', 'info');
                updateInstructions('Preparing your authentication request...');
                
                // Step 1: Get registration options from local server
                const optionsResponse = await fetch(API_URL + '/begin-register?username=' + encodeURIComponent(username));
                if (!optionsResponse.ok) {
                    const error = await optionsResponse.json();
                    throw new Error(error.error || 'Failed to get registration options');
                }
                const registrationOptions = await optionsResponse.json();
                console.log('Registration options:', registrationOptions);
                
                updateStatus('Please interact with your authenticator...', 'processing');
                updateInstructions('You can use: 1) This device\'s biometrics, 2) A security key, or 3) Your phone via QR code (if prompted)');
                
                // Step 2: Use SimpleWebAuthn to create credential
                const credential = await window.SimpleWebAuthnBrowser.startRegistration(registrationOptions);
                console.log('Created credential:', credential);
                
                // Step 3: Send credential to local server to complete registration
                updateStatus('Completing registration...', 'processing');
                const finishResponse = await fetch(API_URL + '/finish-register?username=' + encodeURIComponent(username), {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(credential)
                });
                
                if (!finishResponse.ok) {
                    const error = await finishResponse.json();
                    throw new Error(error.error || 'Failed to complete registration');
                }
                
                const result = await finishResponse.json();
                console.log('Registration result:', result);
                
                // Clear the countdown timer
                clearInterval(countdownInterval);
                
                if (result.success) {
                    updateStatus('Registration successful!', 'success');
                    updateInstructions('Your passkey has been registered. Credential ID: ' + (result.credentialId || 'Created') + '. You can now close this window and return to the CLI.');
                    updateTimeout(0);
                    
                    // Store credential ID if provided
                    if (result.credentialId) {
                        sessionStorage.setItem('sonr_credential_id', result.credentialId);
                    }
                } else {
                    throw new Error(result.error || 'Registration failed');
                }
                
            } catch (error) {
                // Clear the countdown timer
                clearInterval(countdownInterval);
                
                console.error('Registration failed:', error);
                
                // Provide more specific error messages
                let errorMessage = error.message;
                if (error.name === 'NotAllowedError') {
                    errorMessage = 'Registration was cancelled or not allowed';
                } else if (error.name === 'InvalidStateError') {
                    errorMessage = 'An authenticator is already registered';
                } else if (error.name === 'NotSupportedError') {
                    errorMessage = 'This authenticator is not supported';
                }
                
                updateStatus('Registration failed', 'error');
                updateInstructions(errorMessage);
                updateTimeout(0);
            }
        }

        // Start registration when page loads
        window.addEventListener('load', () => {
            // Add a small delay to ensure all resources are loaded
            setTimeout(startRegistration, 500);
        });
    </script>
</body>
</html>`

// webAuthnLoginHTML contains the HTML template for WebAuthn login
const webAuthnLoginHTML = `<!DOCTYPE html>
<html class="dark">
<head>
    <title>Sonr Local Login</title>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <script src="https://cdn.jsdelivr.net/npm/@tailwindcss/browser@4"></script>
    <style>
        :root {
            --sonr-primary: #17c2ff;
            --sonr-primary-hover: #0ea5e9;
        }
    </style>
</head>
<body class="min-h-screen bg-gray-900 flex items-center justify-center p-4">
    <div class="bg-gray-800 p-8 rounded-2xl shadow-2xl max-w-md w-full">
        <div class="text-center mb-8">
            <div class="inline-flex items-center justify-center w-16 h-16 bg-gradient-to-br from-blue-500 to-cyan-400 rounded-full mb-4">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z" />
                </svg>
            </div>
            <h1 class="text-3xl font-bold text-white mb-2">Welcome Back to Sonr</h1>
            <p class="text-gray-400">Authenticating as: <span id="login-username-display" class="font-semibold text-cyan-400">{{.Username}}</span></p>
        </div>

        <div id="status-container" class="mb-6">
            <div id="status" class="p-4 rounded-lg bg-blue-900/50 text-blue-300 text-sm font-medium">
                Initializing WebAuthn authentication...
            </div>
        </div>

        <div id="instructions" class="text-center text-gray-300 mb-6">
            Use your passkey or security key to authenticate.
        </div>

        <div id="timeout-container" class="text-center text-sm text-gray-500">
            <span id="timeout-text"></span>
        </div>
    </div>

    <!-- Load SimpleWebAuthn for WebAuthn operations -->
    <script src="https://unpkg.com/@simplewebauthn/browser@9.0.1/dist/bundle/index.umd.min.js"></script>
    <!-- Load @sonr.io/es for presets and utilities -->
    <script type="module" src="https://unpkg.com/@sonr.io/es@latest/dist/autoloader.js"></script>
    
    <script>
        const TIMEOUT_SECONDS = 30;
        // Support both username and identifier parameters
        const urlParams = new URLSearchParams(window.location.search);
        const username = urlParams.get('identifier') || urlParams.get('username') || "{{.Username}}";
        const rpId = "{{.RPID}}";
        const rpName = "{{.RPName}}";
        const API_URL = window.location.origin; // Use current origin as API URL
        
        // Update the username display if it's from URL params
        if (urlParams.get('identifier') || urlParams.get('username')) {
            document.getElementById('login-username-display').textContent = username;
        }

        function updateStatus(message, type = 'info') {
            const statusEl = document.getElementById('status');
            statusEl.textContent = message;
            
            statusEl.className = 'p-4 rounded-lg text-sm font-medium ';
            if (type === 'success') {
                statusEl.className += 'bg-green-900/50 text-green-300';
            } else if (type === 'error') {
                statusEl.className += 'bg-red-900/50 text-red-300';
            } else {
                statusEl.className += 'bg-blue-900/50 text-blue-300';
            }
        }

        function updateInstructions(text) {
            document.getElementById('instructions').textContent = text;
        }

        function updateTimeout(seconds) {
            const timeoutEl = document.getElementById('timeout-text');
            if (seconds > 0) {
                timeoutEl.textContent = 'Authentication will timeout in ' + seconds + ' seconds';
            } else {
                timeoutEl.textContent = '';
            }
        }

        async function startLogin() {
            let countdownInterval;
            let remainingSeconds = TIMEOUT_SECONDS;

            try {
                // Wait for Sonr to be ready
                await new Promise((resolve) => {
                    if (window.Sonr && window.Sonr.initialized) {
                        resolve();
                    } else {
                        window.addEventListener('sonr:ready', resolve);
                        // Timeout after 5 seconds
                        setTimeout(() => {
                            if (window.Sonr) resolve();
                            else throw new Error('Failed to load Sonr library');
                        }, 5000);
                    }
                });
                
                if (!window.Sonr) {
                    throw new Error('Failed to load Sonr authentication library');
                }
                
                // Check WebAuthn support
                const support = await window.Sonr.webauthn.checkSupport();
                if (!support.supported) {
                    throw new Error('WebAuthn is not supported in this browser. Please use a modern browser like Chrome, Firefox, Safari, or Edge.');
                }
                
                if (!support.platformAuthenticator) {
                    updateStatus('Platform authenticator not available', 'info');
                    updateInstructions('You can still use a security key or phone-based passkey to authenticate.');
                }
                
                updateStatus('Initializing passkey authentication...', 'info');
                updateInstructions('Preparing your authentication request...');

                // Start countdown timer
                countdownInterval = setInterval(() => {
                    remainingSeconds--;
                    updateTimeout(remainingSeconds);
                    if (remainingSeconds <= 0) {
                        clearInterval(countdownInterval);
                        updateStatus('Authentication timed out', 'error');
                        updateInstructions('Please refresh the page to try again.');
                    }
                }, 1000);
                updateTimeout(remainingSeconds);
                
                // Hybrid approach: Use Sonr presets but local server endpoints
                updateStatus('Preparing authentication...', 'info');
                
                // Step 1: Get authentication options from local server
                const optionsResponse = await fetch(API_URL + '/begin-login?username=' + encodeURIComponent(username));
                if (!optionsResponse.ok) {
                    const error = await optionsResponse.json();
                    throw new Error(error.error || 'Failed to get authentication options');
                }
                const authOptions = await optionsResponse.json();
                console.log('Authentication options:', authOptions);
                
                updateStatus('Waiting for your passkey authentication...', 'info');
                updateInstructions('Use your saved passkey from: 1) This device, 2) A security key, or 3) Your phone');
                
                // Step 2: Use SimpleWebAuthn to authenticate
                const credential = await window.SimpleWebAuthnBrowser.startAuthentication(authOptions);
                console.log('Authentication credential:', credential);
                
                // Step 3: Send credential to local server to complete authentication
                updateStatus('Verifying authentication...', 'info');
                const finishResponse = await fetch(API_URL + '/finish-login?username=' + encodeURIComponent(username), {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(credential)
                });
                
                if (!finishResponse.ok) {
                    const error = await finishResponse.json();
                    throw new Error(error.error || 'Failed to complete authentication');
                }
                
                const result = await finishResponse.json();
                console.log('Authentication result:', result);

                // Clear the countdown timer
                clearInterval(countdownInterval);

                if (result.success) {
                    updateStatus('Authentication successful!', 'success');
                    updateInstructions('Welcome back! Credential ID: ' + (result.credentialId || 'Authenticated') + '. You can close this window and return to the CLI.');
                    updateTimeout(0);
                    
                    // Store credential ID if provided
                    if (result.credentialId) {
                        sessionStorage.setItem('sonr_credential_id', result.credentialId);
                    }
                } else {
                    throw new Error(result.error || 'Authentication failed');
                }

            } catch (error) {
                // Clear the countdown timer
                clearInterval(countdownInterval);
                
                console.error('Authentication failed:', error);
                
                // Provide more specific error messages
                let errorMessage = error.message;
                if (error.name === 'NotAllowedError') {
                    errorMessage = 'Authentication was cancelled or not allowed';
                } else if (error.name === 'InvalidStateError') {
                    errorMessage = 'No matching credential found';
                } else if (error.name === 'NotSupportedError') {
                    errorMessage = 'This authenticator is not supported';
                }
                
                updateStatus('Authentication failed', 'error');
                updateInstructions(errorMessage);
                updateTimeout(0);
            }
        }

        // Start login when page loads
        window.addEventListener('load', () => {
            // Add a small delay to ensure all resources are loaded
            setTimeout(startLogin, 500);
        });
    </script>
</body>
</html>`
