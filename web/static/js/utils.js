// CJAdmin 工具函数

// API基础URL
const API_BASE_URL = '/api';

// 获取token
function getToken() {
    return localStorage.getItem('token');
}

// 设置token
function setToken(token) {
    localStorage.setItem('token', token);
}

// 清除token
function clearToken() {
    localStorage.removeItem('token');
}

// 检查是否已登录
function isLoggedIn() {
    return !!getToken();
}

// API请求封装
async function apiRequest(url, options = {}) {
    const token = getToken();
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers
    };

    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }

    const config = {
        ...options,
        headers
    };

    try {
        const response = await fetch(API_BASE_URL + url, config);
        const data = await response.json();

        if (response.status === 401) {
            // 未授权，清除token并跳转到登录页
            clearToken();
            window.location.href = '/login.html';
            throw new Error('未授权，请重新登录');
        }

        if (!response.ok) {
            throw new Error(data.message || '请求失败');
        }

        return data;
    } catch (error) {
        console.error('API请求错误:', error);
        showToast(error.message, 'error');
        throw error;
    }
}

// GET请求
function apiGet(url, params = {}) {
    const queryString = new URLSearchParams(params).toString();
    const fullUrl = queryString ? `${url}?${queryString}` : url;
    return apiRequest(fullUrl, { method: 'GET' });
}

// POST请求
function apiPost(url, data) {
    return apiRequest(url, {
        method: 'POST',
        body: JSON.stringify(data)
    });
}

// PUT请求
function apiPut(url, data) {
    return apiRequest(url, {
        method: 'PUT',
        body: JSON.stringify(data)
    });
}

// DELETE请求
function apiDelete(url) {
    return apiRequest(url, { method: 'DELETE' });
}

// 显示提示消息
function showToast(message, type = 'info') {
    // 移除已存在的toast
    const existingToast = document.querySelector('.toast');
    if (existingToast) {
        existingToast.remove();
    }

    // 创建toast元素
    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    toast.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        padding: 12px 20px;
        border-radius: 4px;
        box-shadow: 0 2px 12px rgba(0,0,0,0.15);
        z-index: 9999;
        animation: slideInRight 0.3s ease-out;
        min-width: 200px;
        max-width: 400px;
    `;

    // 根据类型设置样式
    const colors = {
        success: { bg: '#f0f9ff', color: '#67C23A', border: '#b3e19d' },
        error: { bg: '#fef0f0', color: '#F56C6C', border: '#fbc4c4' },
        warning: { bg: '#fdf6ec', color: '#E6A23C', border: '#f5dab1' },
        info: { bg: '#f4f4f5', color: '#909399', border: '#d3d4d6' }
    };

    const style = colors[type] || colors.info;
    toast.style.backgroundColor = style.bg;
    toast.style.color = style.color;
    toast.style.border = `1px solid ${style.border}`;

    toast.textContent = message;
    document.body.appendChild(toast);

    // 3秒后自动移除
    setTimeout(() => {
        toast.style.animation = 'slideOutRight 0.3s ease-out';
        setTimeout(() => toast.remove(), 300);
    }, 3000);
}

// 确认对话框
function confirm(message) {
    return new Promise((resolve) => {
        const result = window.confirm(message);
        resolve(result);
    });
}

// 格式化日期
function formatDate(dateString, format = 'YYYY-MM-DD HH:mm:ss') {
    if (!dateString) return '-';

    const date = new Date(dateString);
    const year = date.getFullYear();
    const month = String(date.getMonth() + 1).padStart(2, '0');
    const day = String(date.getDate()).padStart(2, '0');
    const hours = String(date.getHours()).padStart(2, '0');
    const minutes = String(date.getMinutes()).padStart(2, '0');
    const seconds = String(date.getSeconds()).padStart(2, '0');

    return format
        .replace('YYYY', year)
        .replace('MM', month)
        .replace('DD', day)
        .replace('HH', hours)
        .replace('mm', minutes)
        .replace('ss', seconds);
}

// 格式化金额
function formatMoney(amount) {
    if (amount === null || amount === undefined) return '-';
    return '¥' + Number(amount).toFixed(2).replace(/\B(?=(\d{3})+(?!\d))/g, ',');
}

// 获取状态徽章HTML
function getStatusBadge(status, statusMap) {
    const info = statusMap[status] || { text: status, type: 'info' };
    return `<span class="badge badge-${info.type}">${info.text}</span>`;
}

// 分页组件
class Pagination {
    constructor(container, options = {}) {
        this.container = typeof container === 'string'
            ? document.querySelector(container)
            : container;
        this.page = options.page || 1;
        this.pageSize = options.pageSize || 10;
        this.total = options.total || 0;
        this.onChange = options.onChange || (() => {});

        this.render();
    }

    setTotal(total) {
        this.total = total;
        this.render();
    }

    render() {
        const totalPages = Math.ceil(this.total / this.pageSize);
        const hasPrev = this.page > 1;
        const hasNext = this.page < totalPages;

        this.container.innerHTML = `
            <div class="pagination">
                <button ${!hasPrev ? 'disabled' : ''} onclick="pagination.prevPage()">上一页</button>
                <span class="page-info">第 ${this.page} / ${totalPages} 页，共 ${this.total} 条</span>
                <button ${!hasNext ? 'disabled' : ''} onclick="pagination.nextPage()">下一页</button>
            </div>
        `;
    }

    prevPage() {
        if (this.page > 1) {
            this.page--;
            this.onChange(this.page, this.pageSize);
            this.render();
        }
    }

    nextPage() {
        const totalPages = Math.ceil(this.total / this.pageSize);
        if (this.page < totalPages) {
            this.page++;
            this.onChange(this.page, this.pageSize);
            this.render();
        }
    }
}

// 表单验证
function validateForm(formData, rules) {
    const errors = [];

    for (const [field, rule] of Object.entries(rules)) {
        const value = formData[field];

        if (rule.required && !value) {
            errors.push(`${rule.label}不能为空`);
            continue;
        }

        if (rule.minLength && value && value.length < rule.minLength) {
            errors.push(`${rule.label}长度不能少于${rule.minLength}个字符`);
        }

        if (rule.maxLength && value && value.length > rule.maxLength) {
            errors.push(`${rule.label}长度不能超过${rule.maxLength}个字符`);
        }

        if (rule.pattern && value && !rule.pattern.test(value)) {
            errors.push(`${rule.label}格式不正确`);
        }

        if (rule.custom && !rule.custom(value)) {
            errors.push(rule.customMessage || `${rule.label}验证失败`);
        }
    }

    return errors;
}

// 导出数据为CSV
function exportToCSV(data, filename) {
    const csv = data.map(row => row.join(',')).join('\n');
    const blob = new Blob(['\ufeff' + csv], { type: 'text/csv;charset=utf-8;' });
    const link = document.createElement('a');
    link.href = URL.createObjectURL(blob);
    link.download = filename;
    link.click();
}

// 模态框管理
class Modal {
    constructor(id) {
        this.modal = document.getElementById(id);
        this.closeBtn = this.modal.querySelector('.modal-close');
        this.overlay = this.modal;

        this.closeBtn.addEventListener('click', () => this.hide());
        this.overlay.addEventListener('click', (e) => {
            if (e.target === this.overlay) {
                this.hide();
            }
        });
    }

    show() {
        this.modal.classList.add('show');
    }

    hide() {
        this.modal.classList.remove('show');
    }
}

// 添加CSS动画
const style = document.createElement('style');
style.textContent = `
    @keyframes slideInRight {
        from {
            transform: translateX(100%);
            opacity: 0;
        }
        to {
            transform: translateX(0);
            opacity: 1;
        }
    }
    @keyframes slideOutRight {
        from {
            transform: translateX(0);
            opacity: 1;
        }
        to {
            transform: translateX(100%);
            opacity: 0;
        }
    }
`;
document.head.appendChild(style);

// 退出登录
function logout() {
    clearToken();
    window.location.href = '/login.html';
}

// 获取URL参数
function getUrlParam(name) {
    const params = new URLSearchParams(window.location.search);
    return params.get(name);
}
