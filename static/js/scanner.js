// State Management
const state = {
    mode: null,           // 'item' или 'location'
    token: null,
    currentUser: null,
    scannedItem: null,
    scannedLocation: null,
    itemInfo: null,
    cameraStream: null,
    isScanning: false,
    recentMoves: []
};

const API_URL = '/api';

// ============================================================================
// AUTHENTICATION
// ============================================================================

async function login() {
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;
    const messageDiv = document.getElementById('authMessage');

    if (!username || !password) {
        showMessage(messageDiv, 'Введите логин и пароль', 'error');
        return;
    }

    try {
        const response = await fetch(`${API_URL}/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ username, password })
        });

        const data = await response.json();

        if (data.success) {
            state.token = data.token;
            state.currentUser = data.username;
            
            // Переключаемся на сканер
            document.getElementById('authSection').style.display = 'none';
            document.getElementById('scannerSection').style.display = 'block';
            document.getElementById('currentUser').textContent = data.username;
            
            showMessage(messageDiv, 'Успешная авторизация! ✓', 'success');
        } else {
            showMessage(messageDiv, data.error || 'Ошибка авторизации', 'error');
        }
    } catch (error) {
        showMessage(messageDiv, 'Ошибка подключения: ' + error.message, 'error');
    }
}

function logout() {
    state.token = null;
    state.currentUser = null;
    state.scannedItem = null;
    state.scannedLocation = null;
    state.itemInfo = null;
    
    stopCamera();
    
    document.getElementById('scannerSection').style.display = 'none';
    document.getElementById('authSection').style.display = 'block';
    
    resetUI();
}

// ============================================================================
// CAMERA & SCANNER
// ============================================================================

async function startCamera() {
    try {
        state.cameraStream = await navigator.mediaDevices.getUserMedia({
            video: {
                facingMode: 'environment',
                width: { ideal: 1280 },
                height: { ideal: 720 }
            },
            audio: false
        });

        const videoElement = document.getElementById('cameraFeed');
        videoElement.srcObject = state.cameraStream;
        await videoElement.play();
        document.getElementById('startButton').style.display = 'none';
        document.getElementById('stopButton').style.display = 'block';
        
        state.isScanning = true;
        updateStatus('📷 Камера открыта - наведите на QR код');
        
        scanQRCode();
    } catch (error) {
        showMessage(
            document.getElementById('message'),
            'Ошибка доступа к камере: ' + error.message,
            'error'
        );
    }
}

function stopCamera() {
    if (state.cameraStream) {
        state.cameraStream.getTracks().forEach(track => track.stop());
        state.cameraStream = null;
    }
    
    state.isScanning = false;
    document.getElementById('startButton').style.display = 'block';
    document.getElementById('stopButton').style.display = 'none';
    updateStatus('Сканирование остановлено');
}

function scanQRCode() {
    if (!state.isScanning) return;

    const video = document.getElementById('cameraFeed');
    const canvas = document.getElementById('scanner');
    const context = canvas.getContext('2d');

    // Устанавливаем размер canvas под размер видео
    const videoWidth = video.videoWidth;
    const videoHeight = video.videoHeight;

    if (videoWidth > 0 && videoHeight > 0) {
        canvas.width = videoWidth;
        canvas.height = videoHeight;
        context.drawImage(video, 0, 0, videoWidth, videoHeight);

        // Получаем пиксели и парсим QR код
        const imageData = context.getImageData(0, 0, videoWidth, videoHeight);
        const code = jsQR(imageData.data, videoWidth, videoHeight);

        if (code) {
            handleQRScan(code.data);
        }
    }

    // Продолжаем сканирование
    requestAnimationFrame(scanQRCode);
}

// ============================================================================
// QR HANDLING
// ============================================================================

async function handleQRScan(qrContent) {
    updateStatus('✓ QR код обнаружен: ' + qrContent);
    
    // Парсим QR: ITEM:item123 или LOC:location7
    if (qrContent.startsWith('ITEM:')) {
        const itemId = qrContent.substring(5);
        await handleItemScan(itemId);
    } else if (qrContent.startsWith('LOC:')) {
        const locationId = qrContent.substring(4);
        await handleLocationScan(locationId);
    } else {
        updateStatus('❌ Неизвестный формат QR: ' + qrContent);
    }
}

async function handleItemScan(itemId) {
    try {
        // Получаем информацию о товаре
        const response = await fetch(`${API_URL}/item/${itemId}`);
        const data = await response.json();

        if (data.success) {
            state.scannedItem = itemId;
            state.itemInfo = data.item;
            
            document.getElementById('scannedItem').textContent = itemId;
            document.getElementById('itemName').textContent = data.item.name;
            document.getElementById('itemSku').textContent = data.item.sku;
            document.getElementById('itemQuantity').textContent = data.item.quantity;
            document.getElementById('itemLocation').textContent = data.item.location?.code || '—';
            
            document.getElementById('itemInfoContainer').style.display = 'block';
            
            updateStatus('✓ Товар отсканирован: ' + data.item.name);
            
            // Если уже есть локация - показываем форму подтверждения
            if (state.scannedLocation) {
                showConfirmation();
            } else {
                updateStatus('👉 Теперь отсканируйте целевую локацию');
            }
        } else {
            updateStatus('❌ Товар не найден: ' + itemId);
        }
    } catch (error) {
        updateStatus('❌ Ошибка получения товара: ' + error.message);
    }
}

async function handleLocationScan(locationId) {
    state.scannedLocation = locationId;
    document.getElementById('scannedLocation').textContent = locationId;
    
    updateStatus('✓ Локация отсканирована: ' + locationId);
    
    // Если уже есть товар - показываем форму подтверждения
    if (state.scannedItem) {
        showConfirmation();
    } else {
        updateStatus('👉 Теперь отсканируйте товар');
    }
}

// ============================================================================
// MODE SWITCHING
// ============================================================================

function switchMode(mode) {
    state.mode = mode;
    
    // Обновляем активные кнопки
    document.querySelectorAll('.btn-mode').forEach(btn => {
        btn.classList.remove('active');
    });
    
    if (mode === 'item') {
        document.getElementById('modeItem').classList.add('active');
    } else if (mode === 'location') {
        document.getElementById('modeLocation').classList.add('active');
    }
    
    if (mode === 'scan') {
        startCamera();
    } else {
        // Не закрываем камеру как раньше - пользователь управляет кноплями
    }
}

// ============================================================================
// CONFIRMATION & SUBMISSION
// ============================================================================

function showConfirmation() {
    if (state.scannedItem && state.scannedLocation) {
        document.getElementById('confirmItem').textContent = state.scannedItem + 
            ' (' + state.itemInfo.name + ')';
        document.getElementById('confirmFromLocation').textContent = 
            state.itemInfo.location?.code || '—';
        document.getElementById('confirmToLocation').textContent = state.scannedLocation;
        
        document.getElementById('confirmContainer').style.display = 'block';
        updateStatus('⚠️ Подтвердите перемещение товара');
    }
}

async function confirmMove() {
    if (!state.scannedItem || !state.scannedLocation) {
        showMessage(document.getElementById('message'), 'Ошибка: отсутствуют данные', 'error');
        return;
    }

    const notes = document.getElementById('notes').value;

    try {
        const response = await fetch(`${API_URL}/move`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({
                item_id: state.scannedItem,
                to_location_id: state.scannedLocation,
                user_id: 'user1', // В реальности из токена
                notes: notes
            })
        });

        const data = await response.json();

        if (data.success) {
            // Добавляем в историю
            const moveRecord = {
                time: new Date().toLocaleTimeString('ru-RU'),
                item: state.scannedItem,
                from: state.itemInfo.location?.code || '—',
                to: state.scannedLocation,
                notes: notes
            };
            state.recentMoves.push(moveRecord);

            showMessage(document.getElementById('message'), 
                '✓ Товар успешно перемещён!', 'success');
            
            updateRecentMoves();
            resetScan();
        } else {
            showMessage(document.getElementById('message'), 
                data.error || 'Ошибка при перемещении товара', 'error');
        }
    } catch (error) {
        showMessage(document.getElementById('message'), 
            'Ошибка: ' + error.message, 'error');
    }
}

function resetScan() {
    state.scannedItem = null;
    state.scannedLocation = null;
    state.itemInfo = null;
    
    document.getElementById('scannedItem').textContent = '—';
    document.getElementById('scannedLocation').textContent = '—';
    document.getElementById('itemInfoContainer').style.display = 'none';
    document.getElementById('confirmContainer').style.display = 'none';
    document.getElementById('notes').value = '';
    
    updateStatus('Готов к новому сканированию');
}

// ============================================================================
// UI HELPERS
// ============================================================================

function updateStatus(message) {
    document.getElementById('currentStatus').textContent = message;
}

function showMessage(element, message, type) {
    element.textContent = message;
    element.className = 'message ' + type;
    
    if (type !== 'error') {
        setTimeout(() => {
            element.className = 'message';
        }, 5000);
    }
}

function updateRecentMoves() {
    const container = document.getElementById('recentMovesContainer');
    const list = document.getElementById('movesList');

    if (state.recentMoves.length > 0) {
        container.style.display = 'block';
        list.innerHTML = '';

        state.recentMoves.slice().reverse().forEach(move => {
            const moveDiv = document.createElement('div');
            moveDiv.className = 'move-item';
            moveDiv.innerHTML = `
                <div class="move-time">${move.time}</div>
                <div class="move-details">
                    <strong>${move.item}</strong>: 
                    ${move.from} → ${move.to}
                    ${move.notes ? `<br/><em>${move.notes}</em>` : ''}
                </div>
            `;
            list.appendChild(moveDiv);
        });
    }
}

function resetUI() {
    document.getElementById('scannedItem').textContent = '—';
    document.getElementById('scannedLocation').textContent = '—';
    document.getElementById('itemInfoContainer').style.display = 'none';
    document.getElementById('confirmContainer').style.display = 'none';
    document.getElementById('recentMovesContainer').style.display = 'none';
    document.getElementById('message').className = 'message';
    document.getElementById('currentStatus').textContent = 'Ожидание сканирования...';
}

// ============================================================================
// INITIALIZATION
// ============================================================================

document.addEventListener('DOMContentLoaded', function() {
    // По умолчанию включаем режим сканирования товара
    switchMode('item');
});

// Закрываем камеру при закрытии страницы
window.addEventListener('unload', function() {
    stopCamera();
});
