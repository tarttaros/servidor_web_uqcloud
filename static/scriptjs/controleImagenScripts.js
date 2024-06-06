var ventanaConfiguracionAbierta = false;

document.addEventListener('DOMContentLoaded', function () {
    const form = document.getElementById('searchForm');
    form.addEventListener('submit', function (event) {
        event.preventDefault();

        const formData = new FormData(form);
        fetch('/api/imagenes', {
            method: 'POST',
            body: formData
        })
            .then(response => response.json())
            .then(data => {
                const tableBody = document.getElementById('imagen-table').getElementsByTagName('tbody')[0];
                tableBody.innerHTML = ''; // Limpiar tabla antes de insertar nuevas filas

                data.forEach(function (imagen) {
                    $("#imagen-table tbody").append(
                        `<tr>
                            <td>${imagen.Repositorio}</td>
                            <td>${imagen.Tag}</td>
                            <td>${imagen.Tamanio}</td>

                            <td class="button-column">
                            <form method="post" action="/eliminarImagen" style="display: inline-block; padding: 0; margin: 0; border: none;">
                            <input type="hidden" id="selectedMachineImagen" name="selectedMachineImagen" value="${imagen.MaquinaVM}">
                            <input type="hidden" name="imagenRepositorio" value="${imagen.Repositorio}">
                            <button type="submit" class="btn btn-link" style="padding: 0; margin: 0;">
                                <img style="width: 35px;" src="/static/images/icons/delete.png" alt="Botón 1">
                            </button>
                            </form>
                            <button type="button" class="btn btn-link" style="padding: 0; margin: 0;" onclick="abrirVentanaEmergenteInformacion('${imagen.Repositorio}','${imagen.Tag}','${imagen.ImagenId}', '${imagen.Creacion}','${imagen.Tamanio}','${imagen.MaquinaVM}')">
                                        <img style="width: 35px;" src="/static/images/icons/info.png" alt="Botón 4">
                            </button>
                            </td>

                        </tr>
                        `
                    );
                });
            })
            .catch(error => console.error('Error fetching data:', error));
    });
});

function abrirVentanaEmergenteInformacion(Repositorio, Tag, ImagenId, Creacion, Tamanio, MaquinaVM) {
    ventanaConfiguracionAbierta = true;

    var ventanaEmergente = document.getElementById("VentanaEmergenteInformacion");
    ventanaEmergente.style.display = "block";

    // Asignar valores a los elementos <span> con sus respectivos ids

    // Nombre de la Máquina
    document.getElementById("repositorioSpan").textContent = Repositorio;

    // Sistema Operativo de la Máquina
    document.getElementById("tagSpan").textContent = Tag;

    // Distribución de la Máquina
    document.getElementById("imagenIdSpan").textContent = ImagenId;

    document.getElementById("creacionSpan").textContent = Creacion;

    // Memoria de la Máquina
    document.getElementById("tamanioSpan").textContent = Tamanio;

    // Estado de la Máquina
    document.getElementById("maquinaVMSpan").textContent = MaquinaVM;
   
}

function cerrarVentanaEmergenteInformacion(){
    ventanaConfiguracionAbierta = false;

    var ventanaEmergente = document.getElementById("VentanaEmergenteInformacion");
    ventanaEmergente.style.display = "none";
}
