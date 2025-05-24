package file_service

import (
    "bytes"
    "io"
    "path"

    shared_models "quickflow/shared/models"
    "quickflow/shared/proto/file_service"
)

func ProtoFileToModel(file *file_service.File) *shared_models.File {
    if file == nil {
        return nil
    }

    return &shared_models.File{
        Name:        file.FileName,
        Size:        file.FileSize,
        MimeType:    file.FileType,
        AccessMode:  shared_models.AccessMode(file.AccessMode),
        URL:         file.Url,
        Reader:      bytes.NewReader(file.File),
        DisplayType: shared_models.DisplayType(file.DisplayType),
        Ext:         path.Ext(file.FileName),
    }
}

func ModelFileToProto(file *shared_models.File) *file_service.File {
    if file == nil {
        return nil
    }

    if file.Reader == nil {
        if len(file.URL) == 0 {
            return nil
        }
        return &file_service.File{
            Url:         file.URL,
            DisplayType: string(file.DisplayType),
            FileName:    file.Name,
        }
    }
    content, err := io.ReadAll(file.Reader)
    if err != nil {
        return nil
    }
    return &file_service.File{
        FileName:    file.Name,
        FileSize:    file.Size,
        FileType:    file.MimeType,
        AccessMode:  file_service.AccessMode(file.AccessMode),
        Url:         file.URL,
        File:        content,
        DisplayType: string(file.DisplayType),
    }
}

func ProtoFilesToModels(files []*file_service.File) []*shared_models.File {
    res := make([]*shared_models.File, len(files))
    for i, file := range files {
        res[i] = ProtoFileToModel(file)
    }
    return res
}

func ModelFilesToProto(files []*shared_models.File) []*file_service.File {
    res := make([]*file_service.File, len(files))
    for i, file := range files {
        res[i] = ModelFileToProto(file)
    }
    return res
}
