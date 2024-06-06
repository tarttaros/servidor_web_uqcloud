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
            .then(responseJson => {
                const contenedores = responseJson.contenedores;
                const imagenes = responseJson.imagen;

                if (Array.isArray(contenedores)) {
                    // Actualizar la tabla con los contenedores
                    const tableBody = document.getElementById('contenedor-table').getElementsByTagName('tbody')[0];
                    tableBody.innerHTML = ''; // Limpiar la tabla antes de insertar nuevas filas

                    contenedores.forEach(function (contenedor) {
                        $("#contenedor-table tbody").append(
                            `<tr>
                            <td>${contenedor.Nombre}</td>
                            <td>${contenedor.Imagen}</td>
                            <td>${contenedor.Status}</td>

                            <td class="button-column">
                            <form method="post" action="/CorrerContenedor" style="display: inline-block; padding: 0; margin: 0; border: none;">
                            <input type="hidden" id="selectedMachineContenedor" name="selectedMachineContenedor" value="${contenedor.MaquinaVM}">
                            <input type="hidden" name="IdContenedor" value="${contenedor.ConetendorId}">
                            <button type="submit" class="btn btn-link" style="padding: 0; margin: 0;">
                                <img style="width: 35px;" src="/static/images/icons/power.png" alt="Botón 1">
                            </button>
                            </form>
                            <form method="post" action="/PausarContenedor" style="display: inline-block; padding: 0; margin: 0; border: none;">
                            <input type="hidden" id="selectedMachineContenedor" name="selectedMachineContenedor" value="${contenedor.MaquinaVM}">
                            <input type="hidden" name="IdContenedor" value="${contenedor.ConetendorId}">
                            <button type="submit" class="btn btn-link" style="padding: 0; margin: 0;">
                                <img style="width: 35px;" src="/static/images/icons/stop.png" alt="Botón 1">
                            </button>
                            </form>
                            <form method="post" action="/ReiniciarContenedor" style="display: inline-block; padding: 0; margin: 0; border: none;">
                            <input type="hidden" id="selectedMachineContenedor" name="selectedMachineContenedor" value="${contenedor.MaquinaVM}">
                            <input type="hidden" name="IdContenedor" value="${contenedor.ConetendorId}">
                            <button type="submit" class="btn btn-link" style="padding: 0; margin: 0;">
                                <img style="width: 35px;" src="/static/images/icons/restart.png" alt="Botón 1">
                            </button>
                            </form>
                            <form method="post" action="/EliminarContenedor" style="display: inline-block; padding: 0; margin: 0; border: none;">
                            <input type="hidden" id="selectedMachineContenedor" name="selectedMachineContenedor" value="${contenedor.MaquinaVM}">
                            <input type="hidden" name="IdContenedor" value="${contenedor.ConetendorId}">
                            <button type="submit" class="btn btn-link" style="padding: 0; margin: 0;">
                                <img style="width: 35px;" src="/static/images/icons/delete.png" alt="Botón 1">
                            </button>
                            </form>
                            <button type="button" class="btn btn-link" style="padding: 0; margin: 0;" onclick="abrirVentanaEmergenteInformacionC('${contenedor.ConetendorId}','${contenedor.Imagen}','${contenedor.Creado}','${contenedor.Status}','${contenedor.Puerto}','${contenedor.Nombre}','${contenedor.MaquinaVM}')">
                                        <img style="width: 35px;" src="/static/images/icons/info.png" alt="Botón 4">
                            </button>
                            </td>
                        </tr>`
                        );
                    });
                }

                if (Array.isArray(imagenes)) {

                    // Rellenar el select con la información adicional
                    const imagenSelect = document.getElementById('buscarImagen');
                    imagenSelect.innerHTML = ''; // Limpiar el select antes de añadir nuevas opciones

                    imagenes.forEach(function (imagen) {
                        const option = document.createElement('option');
                        option.id = 1;
                        option.text = imagen.Repositorio;
                        imagenSelect.appendChild(option);
                    });
                    console.log("Imágenes:", imagenes);

                }
            })
            .catch(error => console.error('Error fetching data:', error));
    });
});

function abrirVentanaEmergenteInformacionC(ConetendorId, Imagen, Creado, Status, Puerto,Nombre,MaquinaVM) {
    ventanaConfiguracionAbierta = true;

    var ventanaEmergente = document.getElementById("VentanaEmergenteInformacion");
    ventanaEmergente.style.display = "block";

    // Asignar valores a los elementos <span> con sus respectivos ids

    // Nombre de la Máquina
    document.getElementById("conetendorIdSpan").textContent = ConetendorId;

    // Sistema Operativo de la Máquina
    document.getElementById("imagenSpan").textContent = Imagen;

    document.getElementById("creadoSpan").textContent = Creado;

    // Memoria de la Máquina
    document.getElementById("statusSpan").textContent = Status;

    // Estado de la Máquina
    document.getElementById("puertoSpan").textContent = Puerto;

    // Memoria de la Máquina
    document.getElementById("nombreSpan").textContent = Nombre;

    // Estado de la Máquina
    document.getElementById("maquinaVMSpan").textContent = MaquinaVM;
   
}

function cerrarVentanaEmergenteInformacion(){
    ventanaConfiguracionAbierta = false;

    var ventanaEmergente = document.getElementById("VentanaEmergenteInformacion");
    ventanaEmergente.style.display = "none";
}
