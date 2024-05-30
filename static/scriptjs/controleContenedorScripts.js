document.addEventListener('DOMContentLoaded', function () {
    const form = document.getElementById('searchForm');
    form.addEventListener('submit', function (event) {
        event.preventDefault();

        const formData = new FormData(form);
        fetch('/api/contendores', {
            method: 'POST',
            body: formData
        })
            .then(response => {
                if (!response.ok) {
                    throw new Error('Network response was not ok');
                }
                return response.json();
            })
            .then(data => {
                const tableBody = document.getElementById('contenedor-table').getElementsByTagName('tbody')[0];
                tableBody.innerHTML = ''; // Limpiar tabla antes de insertar nuevas filas

                data.forEach(function(contenedores) {
                    $("#contenedor-table tbody").append(
                        `<tr>
                            <td>${contenedores.ConetendorId}</td>
                            <td>${contenedores.Imagen}</td>
                            <td>${contenedores.Status}</td>

                            <td class="button-column">
                            <form method="post" action="/CorrerContenedor" style="display: inline-block; padding: 0; margin: 0; border: none;">
                            <input type="hidden" id="selectedMachineContenedor" name="selectedMachineContenedor" value="${contenedores.MaquinaVM}">
                            <input type="hidden" name="IdContenedor" value="${contenedores.ConetendorId}">
                            <button type="submit" class="btn btn-link" style="padding: 0; margin: 0;">
                                <img style="width: 35px;" src="/static/images/icons/power.png" alt="Botón 1">
                            </button>
                            </form>
                            <form method="post" action="/PausarContenedor" style="display: inline-block; padding: 0; margin: 0; border: none;">
                            <input type="hidden" id="selectedMachineContenedor" name="selectedMachineContenedor" value="${contenedores.MaquinaVM}">
                            <input type="hidden" name="IdContenedor" value="${contenedores.ConetendorId}">
                            <button type="submit" class="btn btn-link" style="padding: 0; margin: 0;">
                                <img style="width: 35px;" src="/static/images/icons/stop.png" alt="Botón 1">
                            </button>
                            </form>
                            <form method="post" action="/ReiniciarContenedor" style="display: inline-block; padding: 0; margin: 0; border: none;">
                            <input type="hidden" id="selectedMachineContenedor" name="selectedMachineContenedor" value="${contenedores.MaquinaVM}">
                            <input type="hidden" name="IdContenedor" value="${contenedores.ConetendorId}">
                            <button type="submit" class="btn btn-link" style="padding: 0; margin: 0;">
                                <img style="width: 35px;" src="/static/images/icons/restart.png" alt="Botón 1">
                            </button>
                            </form>
                            <form method="post" action="/EliminarContenedor" style="display: inline-block; padding: 0; margin: 0; border: none;">
                            <input type="hidden" id="selectedMachineContenedor" name="selectedMachineContenedor" value="${contenedores.MaquinaVM}">
                            <input type="hidden" name="IdContenedor" value="${contenedores.ConetendorId}">
                            <button type="submit" class="btn btn-link" style="padding: 0; margin: 0;">
                                <img style="width: 35px;" src="/static/images/icons/delete.png" alt="Botón 1">
                            </button>
                            </form>
                            </td>
                        </tr>`
                    );
                });
            })
            .catch(error => console.error('Error fetching data:', error));
    });
});
