// ===== NOTIFICATION SYSTEM =====
class NotificationManager {
    constructor() {
        this.container = this.createContainer();
    }

    createContainer() {
        const container = document.createElement('div');
        container.id = 'notification-container';
        container.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            z-index: 10000;
            display: flex;
            flex-direction: column;
            gap: 10px;
            pointer-events: none;
        `;
        document.body.appendChild(container);
        return container;
    }

    show(message, type = 'info', duration = 3000) {
        const notification = this.createNotification(message, type);
        this.container.appendChild(notification);

        // Trigger animation
        setTimeout(() => notification.classList.add('show'), 10);

        // Auto remove
        if (duration > 0) {
            setTimeout(() => this.remove(notification), duration);
        }

        return notification;
    }

    createNotification(message, type) {
        const colors = {
            success: { bg: '#10b981', border: '#059669', icon: '✓' },
            error: { bg: '#ef4444', border: '#dc2626', icon: '✕' },
            warning: { bg: '#f59e0b', border: '#d97706', icon: '⚠' },
            info: { bg: '#06b6d4', border: '#0891b2', icon: 'ℹ' }
        };

        const style = colors[type] || colors.info;

        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.style.cssText = `
            background: ${style.bg};
            border: 1px solid ${style.border};
            color: white;
            padding: 1rem 1.5rem;
            border-radius: 8px;
            box-shadow: 0 10px 25px rgba(0,0,0,0.3);
            display: flex;
            align-items: center;
            gap: 12px;
            min-width: 300px;
            max-width: 500px;
            pointer-events: all;
            transform: translateX(400px);
            opacity: 0;
            transition: all 0.4s cubic-bezier(0.68, -0.55, 0.265, 1.55);
            cursor: pointer;
        `;

        notification.innerHTML = `
            <span style="font-size: 1.25rem; font-weight: bold;">${style.icon}</span>
            <span style="flex: 1; font-weight: 500;">${message}</span>
            <button onclick="this.closest('.notification').remove()" style="
                background: rgba(255,255,255,0.2);
                border: none;
                color: white;
                width: 24px;
                height: 24px;
                border-radius: 4px;
                cursor: pointer;
                display: flex;
                align-items: center;
                justify-content: center;
                font-size: 14px;
            ">×</button>
        `;

        notification.onclick = (e) => {
            if (e.target.tagName !== 'BUTTON') {
                this.remove(notification);
            }
        };

        return notification;
    }

    remove(notification) {
        notification.style.transform = 'translateX(400px)';
        notification.style.opacity = '0';
        setTimeout(() => notification.remove(), 400);
    }

    success(message, duration) {
        return this.show(message, 'success', duration);
    }

    error(message, duration) {
        return this.show(message, 'error', duration);
    }

    warning(message, duration) {
        return this.show(message, 'warning', duration);
    }

    info(message, duration) {
        return this.show(message, 'info', duration);
    }
}

// ===== LOADING OVERLAY =====
class LoadingOverlay {
    static show(message = 'Подключение...') {
        const overlay = document.createElement('div');
        overlay.id = 'loading-overlay';
        overlay.style.cssText = `
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(10, 14, 23, 0.9);
            backdrop-filter: blur(5px);
            z-index: 9999;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            gap: 20px;
            animation: fadeIn 0.3s ease;
        `;

        overlay.innerHTML = `
            <div class="loader" style="
                width: 60px;
                height: 60px;
                border: 4px solid rgba(6, 182, 212, 0.1);
                border-top-color: #06b6d4;
                border-radius: 50%;
                animation: spin 1s linear infinite;
                box-shadow: 0 0 20px rgba(6, 182, 212, 0.5);
            "></div>
            <p style="color: #e2e8f0; font-size: 1.125rem; font-weight: 500; animation: pulse 2s ease infinite;">${message}</p>
        `;

        document.body.appendChild(overlay);
        document.body.style.overflow = 'hidden';
    }

    static hide() {
        const overlay = document.getElementById('loading-overlay');
        if (overlay) {
            overlay.style.animation = 'fadeOut 0.3s ease';
            setTimeout(() => {
                overlay.remove();
                document.body.style.overflow = '';
            }, 300);
        }
    }
}

// ===== INPUT ANIMATIONS =====
function initInputAnimations() {
    const inputs = document.querySelectorAll('input[type="text"], input[type="password"]');
    
    inputs.forEach(input => {
        // Add floating label effect
        const label = input.previousElementSibling;
        if (label && label.tagName === 'LABEL') {
            label.style.transition = 'all 0.3s ease';
            
            input.addEventListener('focus', () => {
                label.style.color = '#06b6d4';
                label.style.transform = 'translateY(-2px)';
            });
            
            input.addEventListener('blur', () => {
                label.style.color = '';
                label.style.transform = '';
            });
        }

        // Add input glow effect
        input.addEventListener('input', (e) => {
            if (e.target.value.length > 0) {
                e.target.style.boxShadow = '0 0 15px rgba(6, 182, 212, 0.2)';
            } else {
                e.target.style.boxShadow = '';
            }
        });
    });
}

// ===== BUTTON RIPPLE EFFECT =====
function initRippleEffect() {
    const button = document.getElementById('connectBtn');
    if (!button) return;

    button.addEventListener('click', function(e) {
        const rect = button.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;

        const ripple = document.createElement('span');
        ripple.style.cssText = `
            position: absolute;
            width: 20px;
            height: 20px;
            background: rgba(255, 255, 255, 0.5);
            border-radius: 50%;
            transform: translate(-50%, -50%);
            left: ${x}px;
            top: ${y}px;
            animation: ripple 0.6s ease-out;
            pointer-events: none;
        `;

        button.style.position = 'relative';
        button.style.overflow = 'hidden';
        button.appendChild(ripple);

        setTimeout(() => ripple.remove(), 600);
    });
}

// ===== CONFETTI EFFECT =====
function triggerConfetti() {
    const colors = ['#06b6d4', '#3b82f6', '#10b981', '#f59e0b', '#ef4444'];
    const confettiCount = 100;

    for (let i = 0; i < confettiCount; i++) {
        setTimeout(() => {
            const confetti = document.createElement('div');
            confetti.style.cssText = `
                position: fixed;
                width: 10px;
                height: 10px;
                background: ${colors[Math.floor(Math.random() * colors.length)]};
                top: -10px;
                left: ${Math.random() * 100}vw;
                border-radius: ${Math.random() > 0.5 ? '50%' : '0'};
                pointer-events: none;
                z-index: 10001;
                animation: confetti-fall ${2 + Math.random() * 2}s linear forwards;
            `;
            
            document.body.appendChild(confetti);
            setTimeout(() => confetti.remove(), 4000);
        }, i * 10);
    }
}

// ===== ENHANCED FORM VALIDATION =====
function validateField(input) {
    const value = input.value.trim();
    let isValid = true;
    let message = '';

    if (input.hasAttribute('required') && !value) {
        isValid = false;
        message = 'Это поле обязательно для заполнения';
    } else if (input.type === 'email' && value && !isValidEmail(value)) {
        isValid = false;
        message = 'Введите корректный email';
    }

    // Visual feedback
    if (!isValid) {
        input.style.borderColor = '#ef4444';
        input.style.animation = 'shake 0.5s ease';
        setTimeout(() => {
            input.style.animation = '';
        }, 500);
    } else {
        input.style.borderColor = '#10b981';
        setTimeout(() => {
            input.style.borderColor = '';
        }, 1000);
    }

    return { isValid, message };
}

function isValidEmail(email) {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

// ===== INITIALIZATION =====
document.addEventListener('DOMContentLoaded', () => {
    initInputAnimations();
    initRippleEffect();

    // Add validation on blur
    document.querySelectorAll('input[required]').forEach(input => {
        input.addEventListener('blur', () => validateField(input));
    });
});

// ===== ADD CSS ANIMATIONS =====
const style = document.createElement('style');
style.textContent = `
    @keyframes fadeIn {
        from { opacity: 0; }
        to { opacity: 1; }
    }
    
    @keyframes fadeOut {
        from { opacity: 1; }
        to { opacity: 0; }
    }
    
    @keyframes spin {
        to { transform: rotate(360deg); }
    }
    
    @keyframes pulse {
        0%, 100% { opacity: 1; }
        50% { opacity: 0.6; }
    }
    
    @keyframes shake {
        0%, 100% { transform: translateX(0); }
        25% { transform: translateX(-10px); }
        75% { transform: translateX(10px); }
    }
    
    @keyframes ripple {
        to {
            width: 200px;
            height: 200px;
            margin-left: -100px;
            margin-top: -100px;
            opacity: 0;
        }
    }
    
    @keyframes confetti-fall {
        0% {
            transform: translateY(0) rotate(0deg);
            opacity: 1;
        }
        100% {
            transform: translateY(100vh) rotate(720deg);
            opacity: 0;
        }
    }
    
    .notification.show {
        transform: translateX(0) !important;
        opacity: 1 !important;
    }
`;
document.head.appendChild(style);

// ===== EXPORT GLOBAL FUNCTIONS =====
window.notifications = new NotificationManager();
window.LoadingOverlay = LoadingOverlay;
window.triggerConfetti = triggerConfetti;
window.validateField = validateField;