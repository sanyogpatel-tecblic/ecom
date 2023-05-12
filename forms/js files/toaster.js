

    function toastersuccess(message) {
        const notyf = new Notyf({
            duration: 2000,
            position: {
                x: 'right',
                y: 'top',
            },
            types: [
                {
                    type: 'warning',
                    background: 'indianred',
                    icon: {
                        className: 'fas fa-exclamation-circle',
                        tagName: 'span',
                        color: '#FFF',
                        close: true,
                    },
                },
            ],
            dismissible: true,
        });

        notyf.open({
            type: 'warning',
            background: 'rgba(60, 179, 113, 0.9)',
            message: message,
        });
    }
    function toastererror(message) {
        const notyf = new Notyf({
            duration: 2000,
            position: {
                x: 'right',
                y: 'top',
            },
            types: [
                {
                    type: 'warning',
                    background: 'indianred',
                    icon: {
                        className: 'fas fa-exclamation-circle',
                        tagName: 'span',
                        color: '#FFF',
                        close: true,
                    },
                },
            ],
            dismissible: true,
        });

        notyf.open({
            type: 'warning',
            background: 'rgba(255, 99, 71, 0.9)',
            message: message,
        });
    }
